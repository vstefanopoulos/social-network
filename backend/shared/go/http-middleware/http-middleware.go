package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"slices"
	ct "social-network/shared/go/ct"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/jwt"
	tele "social-network/shared/go/telemetry"

	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ratelimiter interface {
	Allow(ctx context.Context, key string, limit int, durationSeconds int64) (bool, error)
}

type middleware struct {
	ratelimiter ratelimiter
	serviceName string
	mux         *http.ServeMux
}

func NewMiddleware(ratelimiter ratelimiter, serviceName string, mux *http.ServeMux) *middleware {
	return &middleware{
		ratelimiter: ratelimiter,
		serviceName: serviceName,
		mux:         mux,
	}
}

// MiddleSystem holds the middleware chain
type MiddleSystem struct {
	ratelimiter ratelimiter
	serviceName string
	endpoint    string
	mux         *http.ServeMux

	middlewareChain []func(http.ResponseWriter, *http.Request) (bool, *http.Request)
	middlewareNames []string
	methods         []string
}

// Chain initializes a new middleware chain
func (m *middleware) SetEndpoint(endpoint string) *MiddleSystem {
	return &MiddleSystem{
		ratelimiter: m.ratelimiter,
		serviceName: m.serviceName,
		endpoint:    endpoint,
		mux:         m.mux,
	}
}

// add appends a middleware function to the chain
func (m *MiddleSystem) add(name string, f func(http.ResponseWriter, *http.Request) (bool, *http.Request)) {
	m.middlewareNames = append(m.middlewareNames, name)
	m.middlewareChain = append(m.middlewareChain, f)
}

// AllowedMethod sets allowed HTTP methods and handles CORS preflight requests
func (m *MiddleSystem) AllowedMethod(methods ...string) *MiddleSystem {
	m.methods = methods
	m.add("allowedMethods", func(w http.ResponseWriter, r *http.Request) (bool, *http.Request) {
		ctx := r.Context()
		tele.Debug(ctx, fmt.Sprint("endpoint called:", r.URL.Path, " with method: ", r.Method))
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", ")+", OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-Id, X-Timestamp, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// TODO fix this, return cors to be
		// w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8081")
		// w.Header().Set("Access-Control-Allow-Origin", "*")
		// w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ", ")+", OPTIONS")
		// w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-Id, X-Timestamp")
		// w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			tele.Debug(ctx, "Method in options")
			w.WriteHeader(http.StatusNoContent) // 204
			return false, nil
		}

		if slices.Contains(methods, r.Method) {
			return true, r
		}

		// method not allowed
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		tele.Warn(ctx, "method not allowed. received @1, @2", "method", r.Method, "required", methods)
		return false, nil
	})
	return m
}

// EnrichContext adds metadata to the request context
func (m *MiddleSystem) EnrichContext() *MiddleSystem {
	m.add("EnrichContext", func(w http.ResponseWriter, r *http.Request) (bool, *http.Request) {
		r = utils.RequestWithValue(r, ct.ReqID, utils.GenUUID())
		r = utils.RequestWithValue(r, ct.IP, r.RemoteAddr)
		return true, r
	})
	return m
}

// Auth middleware to validate JWT and enrich context with claims
func (m *MiddleSystem) Auth() *MiddleSystem {
	m.add("Auth", func(w http.ResponseWriter, r *http.Request) (bool, *http.Request) {
		ctx := r.Context()

		tele.Debug(ctx, "authenticating @1 with @2", "endpoint", r.URL, "cookies", r.Cookies())
		cookie, err := r.Cookie("jwt")
		if err != nil {
			tele.Warn(ctx, "no cookie found @1 @2", "endpoint", r.URL, "error", err.Error())
			utils.ErrorJSON(ctx, w, http.StatusUnauthorized, "missing auth cookie")
			return false, nil
		}
		tele.Debug(ctx, "JWT cookie. @1", "value", cookie.Value)
		claims, err := jwt.ParseAndValidate(cookie.Value)
		if err != nil {
			tele.Warn(ctx, "unauthorized request at @1 @2", "endpoint", r.URL, "error", err.Error())
			utils.ErrorJSON(ctx, w, http.StatusUnauthorized, err.Error())
			return false, nil
		}
		// enrich request with claims
		tele.Debug(ctx, "authorization successfull @1", "endpoint", r.URL)
		r = utils.RequestWithValue(r, ct.ClaimsKey, claims)
		r = utils.RequestWithValue(r, ct.UserId, claims.UserId)
		tele.Debug(ctx, "adding these to context @1 at @2", "claims", claims, "endpoint", r.URL)
		return true, r
	})
	return m
}

// // BindReqMeta binds request metadata to context
// func (m *MiddleSystem) BindReqMeta() *MiddleSystem {
// 	m.add(func(w http.ResponseWriter, r *http.Request) (bool, *http.Request) {
// 		r = utils.RequestWithValue(r, ct.ReqActionDetails, r.Header.Get("X-Action-Details"))
// 		r = utils.RequestWithValue(r, ct.ReqTimestamp, r.Header.Get("X-Timestamp"))
// 		return true, r
// 	})
// 	return m
// }

type rateLimitType struct {
}

var (
	UserLimit = rateLimitType{}
	IPLimit   = rateLimitType{}
)

func (m *MiddleSystem) RateLimit(rateLimitType rateLimitType, limit int, durationSeconds int64) *MiddleSystem {
	m.add("RateLimit", func(w http.ResponseWriter, r *http.Request) (bool, *http.Request) {
		ctx := r.Context()
		tele.Debug(ctx, "in ratelimit of @1 for @2", "type", rateLimitType, "endpoint", r.URL)
		rateLimitKey := ""

		switch rateLimitType {
		case IPLimit:
			remoteIp, err := getRemoteIpKey(r)
			if err != nil {
				tele.Warn(ctx, "malformed @1 found", "ip", remoteIp, "error", err.Error())
				utils.ErrorJSON(ctx, w, http.StatusNotAcceptable, "your IP is absolutely WACK")
				return false, nil
			}
			rateLimitKey = fmt.Sprintf("%s:%s:ip:%s", m.serviceName, m.endpoint, remoteIp)

		case UserLimit:
			userId, ok := ctx.Value(ct.UserId).(int64)
			if !ok {
				tele.Warn(ctx, "err or missing userId. @1 @2", "userId", userId, "found_ctx_value", ok)
				utils.ErrorJSON(ctx, w, http.StatusNotAcceptable, "how the hell did you end up here without a user id?")
				return false, nil
			}
			rateLimitKey = fmt.Sprintf("%s:%s:id:%d", m.serviceName, m.endpoint, userId)

		default:
			tele.Error(ctx, "bad rate limit type used!!")
			panic("bad rate limit type argument!")
		}

		ok, err := m.ratelimiter.Allow(ctx, rateLimitKey, limit, durationSeconds)
		if err != nil {
			tele.Info(ctx, "ratelimiter @1 for @2", "error", err.Error(), "endpoint", r.URL)
			// utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "you broke the rate limiter")
			return true, nil
		}
		if !ok {
			tele.Info(ctx, "rate limited for @1, reached @2 per @3", "key", rateLimitKey, "limit", limit, "seconds", durationSeconds)
			utils.ErrorJSON(ctx, w, http.StatusTooManyRequests, "429: stop it, get some help")
			return false, nil
		}
		return true, r
	})
	return m
}

func getRemoteIpKey(r *http.Request) (string, error) {
	remoteIp, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteIp = r.RemoteAddr
	}
	if remoteIp == "" {
		//ip is broken somehow
		return "", nil
	}
	return remoteIp, nil
}

// Finalize constructs the final http.HandlerFunc with all middleware applied and adds them to the mux
func (m *MiddleSystem) Finalize(next http.HandlerFunc) {
	for _, method := range m.methods {

		//create a custom handler func that runs our custom middleware
		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			newCtx, span := tele.Trace(r.Context(), "http middleware start")

			//swapping the context so that the above trace metadata survives
			r = r.WithContext(newCtx)

			defer func() {
				if rec := recover(); rec != nil {
					tele.Error(r.Context(), "panic occured in @1", r.URL)
				}
			}()

			for i, mw := range m.middlewareChain {
				ctx, span := tele.Trace(r.Context(), "start of middleware step", "stepIndex", i)
				defer span.End()
				r = r.WithContext(ctx)
				proceed, newReq := mw(w, r)
				r = newReq
				if !proceed {
					return
				}
			}

			tele.Info(r.Context(), "middleware finished, calling @1", "endpoint", r.URL)
			span.End()
			next.ServeHTTP(w, r)
		}

		//create this struct, that conforms to the handler interface, which will just trigger our custom handler
		//this is used so that we can use our custom handler function with otel's handler that automated some stuff, like trace loading to context and something about metrics
		applier := handlerApplier{
			handle: handlerFunc,
		}

		//creating the otel handler based on our applier
		otelledHandler := otelhttp.NewHandler(applier, m.endpoint)

		//passing the handler function to the mux
		m.mux.HandleFunc(method+" "+m.endpoint, otelledHandler.ServeHTTP)
	}
}

type handlerApplier struct {
	handle func(http.ResponseWriter, *http.Request)
}

func (a handlerApplier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.handle(w, r)
}
