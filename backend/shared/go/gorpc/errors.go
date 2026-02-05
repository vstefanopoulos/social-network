package gorpc

import (
	"context"
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ClassifiedError struct {
	Class       ErrorClass
	GRPCCode    codes.Code
	Retryable   bool
	Description string
}

type ErrorClass string

const (
	ErrorClassClient       ErrorClass = "CLIENT_ERROR"
	ErrorClassServer       ErrorClass = "SERVER_ERROR"
	ErrorClassRetryable    ErrorClass = "RETRYABLE_DEPENDENCY_ERROR"
	ErrorClassNonRetryable ErrorClass = "NON_RETRYABLE_DEPENDENCY_ERROR"
	ErrorClassTimeout      ErrorClass = "TIMEOUT"
	ErrorClassCanceled     ErrorClass = "CANCELED"
	ErrorClassUnavailable  ErrorClass = "UNAVAILABLE"
	ErrorClassTransport    ErrorClass = "TRANSPORT_ERROR"
	ErrorClassUnknown      ErrorClass = "UNKNOWN_ERROR"
)

func Classify(err error) (httpStatus int, class ClassifiedError) {
	if err == nil {
		return http.StatusOK, ClassifiedError{}
	}

	// ---- Context errors (caller-side) ----
	if errors.Is(err, context.DeadlineExceeded) {
		return GrpcCodeToHTTP(codes.DeadlineExceeded), ClassifiedError{
			Class:       ErrorClassTimeout,
			GRPCCode:    codes.DeadlineExceeded,
			Retryable:   true,
			Description: "request timed out",
		}
	}

	if errors.Is(err, context.Canceled) {
		return GrpcCodeToHTTP(codes.Canceled), ClassifiedError{
			Class:       ErrorClassCanceled,
			GRPCCode:    codes.Canceled,
			Retryable:   false,
			Description: "request canceled",
		}
	}

	// ---- gRPC status errors ----
	st, ok := status.FromError(err)
	if !ok {
		// Non-gRPC error (network / transport)
		return GrpcCodeToHTTP(codes.Unknown), ClassifiedError{
			Class:       ErrorClassTransport,
			GRPCCode:    codes.Unknown,
			Retryable:   true,
			Description: "transport or connection error",
		}
	}

	code := st.Code()

	httpStatus = GrpcCodeToHTTP(code)

	switch code {

	// ---- Client errors (do NOT retry) ----
	case codes.InvalidArgument,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated,
		codes.FailedPrecondition,
		codes.OutOfRange:

		return httpStatus, ClassifiedError{
			Class:       ErrorClassClient,
			GRPCCode:    code,
			Retryable:   false,
			Description: "downstream client error",
		}

	// ---- Server errors (retryable) ----
	case codes.Internal,
		codes.DataLoss,
		codes.Unknown:

		return httpStatus, ClassifiedError{
			Class:       ErrorClassServer,
			GRPCCode:    code,
			Retryable:   true,
			Description: "downstream server error",
		}

	// ---- Availability / dependency failures ----
	case codes.Unavailable:
		return httpStatus, ClassifiedError{
			Class:       ErrorClassUnavailable,
			GRPCCode:    code,
			Retryable:   true,
			Description: "downstream service unavailable",
		}

	case codes.ResourceExhausted:
		return httpStatus, ClassifiedError{
			Class:       ErrorClassRetryable,
			GRPCCode:    code,
			Retryable:   true,
			Description: "downstream resource exhausted",
		}

	case codes.Aborted:
		return httpStatus, ClassifiedError{
			Class:       ErrorClassRetryable,
			GRPCCode:    code,
			Retryable:   true,
			Description: "request aborted, retry recommended",
		}

	case codes.DeadlineExceeded:
		return httpStatus, ClassifiedError{
			Class:       ErrorClassTimeout,
			GRPCCode:    code,
			Retryable:   true,
			Description: "downstream deadline exceeded",
		}

	// ---- Explicit non-retryable dependency errors ----
	case codes.Unimplemented:
		return httpStatus, ClassifiedError{
			Class:       ErrorClassNonRetryable,
			GRPCCode:    code,
			Retryable:   false,
			Description: "method not implemented",
		}

	default:
		return httpStatus, ClassifiedError{
			Class:       ErrorClassUnknown,
			GRPCCode:    code,
			Retryable:   false,
			Description: "unclassified grpc error",
		}
	}
}

func GrpcCodeToHTTP(code codes.Code) int {
	switch code {

	// ---- Client errors ----
	case codes.InvalidArgument,
		codes.OutOfRange:
		return http.StatusBadRequest // 400

	case codes.Unauthenticated:
		return http.StatusUnauthorized // 401

	case codes.PermissionDenied:
		return http.StatusForbidden // 403

	case codes.NotFound:
		return http.StatusNotFound // 404

	case codes.AlreadyExists:
		return http.StatusConflict // 409

	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed // 412

	// ---- Rate limiting / quota ----
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests // 429

	// ---- Timeout ----
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout // 504

	// ---- Dependency / availability ----
	case codes.Unavailable:
		return http.StatusServiceUnavailable // 503

	case codes.Aborted:
		return http.StatusConflict // 409 (safe + common)

	// ---- Server errors ----
	case codes.Internal,
		codes.DataLoss,
		codes.Unknown:
		return http.StatusInternalServerError // 500

	case codes.Unimplemented:
		return http.StatusNotImplemented // 501

	default:
		return http.StatusInternalServerError // 500
	}
}
