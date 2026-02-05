package tele

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"social-network/shared/go/commonerrors"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"
)

var ErrUnevenArgs = errors.New("passed arguments aren't even")

type LogLevel struct {
	tag   string
	level int
}

// TODO
//get stack info
//extra context info
//log the 3 functions that called the log

type logging struct {
	serviceName string
	enableDebug bool     //if debug prints will be shown or not
	contextKeys []string //the context keys that will be added to logs as metadata
	slog        *slog.Logger
	simplePrint bool   //if it should print logs in a simple way, or a super verbose way with all details
	prefix      string //this will be added at the start of logs that appear in the local terminal only, suggestion: keep it 3 letters CAPITAL, ex: API, MED, SOC, NOT, POS, RED, USE, CHA
	hasPrefix   bool
}

// newLogger returns a logger that actually logs, uses a handler that taken from a global provider created by the otel sdk
func newLogger(serviceName string, contextKeys contextKeys, enableDebug bool, simplePrint bool, prefix string) *logging {
	handler := otelslog.NewHandler(
		serviceName,
		otelslog.WithLoggerProvider(global.GetLoggerProvider()),
	)

	logger := slog.New(handler)

	if prefix == "" {
		prefix = serviceName
	}

	logging := &logging{
		serviceName: serviceName,
		contextKeys: contextKeys.GetKeys(),
		slog:        logger,
		enableDebug: enableDebug,
		simplePrint: simplePrint,
		prefix:      prefix,
	}

	return logging
}

func newBasicLogger() *logging {
	return &logging{
		serviceName: "not-initalized",
		slog:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

//TODO find issue with malformed context breaking formatArgs due to out of index [2] 2

func (l *logging) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if level == slog.LevelDebug && l.enableDebug == false {
		return
	}

	if len(args)%2 != 0 {
		Error(ctx, "TELE LOG WITH BAD ARGUMENTS FOUND!!: "+fmt.Sprint(args...))
	}

	argsAlreadyPrinted := make([]bool, (len(args)/2)*2)
	var messageBuilder strings.Builder
	callerInfo := functionCallers()
	var max byte = byte(min(9, len(args)/2))

	for i := 0; i < len(msg); i++ {
		curr := msg[i]
		if curr == '@' {
			if i == len(msg) {
				break
			}
			next := msg[i+1] - '0'
			if next >= 1 && next <= 9 && next <= max {
				i++
				key, ok := args[(next-1)*2].(string)
				if ok == false {
					key = "bad_key!"
				}
				// messageBuilder.WriteRune('(')
				messageBuilder.WriteString(key)
				messageBuilder.WriteString("=")
				val, ok := args[(next-1)*2+1].(string)
				if ok == false {
					val = formatArg(args[(next-1)*2+1])
				}
				messageBuilder.WriteString(val)
				// messageBuilder.WriteRune(')')
				argsAlreadyPrinted[(next-1)*2] = true
				argsAlreadyPrinted[(next-1)*2+1] = true
				continue
			}
		}
		messageBuilder.WriteByte(curr)
	}

	ctxArgs := []slog.Attr{}
	if ctx == nil {
		ctx = context.Background()
	} else {
		ctxArgs = l.context2Attributes(ctx)
	}

	message := messageBuilder.String()
	l.slog.Log(
		ctx,
		level,
		message,
		slog.GroupAttrs("customArgs", kvPairsToAttrs(args)...),
		slog.GroupAttrs("context", ctxArgs...),
		slog.String("callers", callerInfo),
		slog.String("prefix", l.prefix),
	)

	if !l.simplePrint && len(ctxArgs) > 0 {
		args = append(args, ctxArgs)
	}

	messageBuilder.Reset()
	messageBuilder.WriteString(time.Now().Format("15:04:05.000"))
	messageBuilder.WriteString(" [")
	messageBuilder.WriteString(l.prefix)
	messageBuilder.WriteString("]: ")
	messageBuilder.WriteString(level.String())
	messageBuilder.WriteRune(' ')
	messageBuilder.WriteString(message)
	if len(args) > 0 {
		formatArgs(&messageBuilder, argsAlreadyPrinted, args...)
	}
	messageBuilder.WriteRune('\n')
	fmt.Fprint(os.Stdout, messageBuilder.String())
}

func kvPairsToAttrs(pairs []any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			key = "invalid_key"
		}
		if len(pairs) < i+1 {
			attrs = append(attrs, slog.Any(key, pairs[i+1]))
		} else {
			attrs = append(attrs, slog.Any(key, "MISSING VALUE!"))
		}
	}
	return attrs
}

func formatArgs(builder *strings.Builder, alreadyPrinted []bool, args ...any) {
	prefixPrinted := false
	for i, arg := range args {
		if len(alreadyPrinted) > i && alreadyPrinted[i] {
			continue
		}
		if prefixPrinted == false {
			builder.WriteString(" args: ")
			prefixPrinted = true
		}

		v := reflect.ValueOf(arg)

		// Handle pointers
		if v.Kind() == reflect.Pointer && !v.IsNil() {
			v = v.Elem()
		}

		if v.Kind() == reflect.Struct {
			fmt.Fprint(builder, commonerrors.FormatValue(v))
		} else {
			fmt.Fprint(builder, arg)
		}

		if i%2 == 0 {
			builder.WriteString(":")
		} else {
			builder.WriteString(" ")
		}

	}

}

func formatArg(arg any) string {
	// v := reflect.ValueOf(arg)

	// // Handle pointers
	// if v.Kind() == reflect.Pointer && !v.IsNil() {
	// 	v = v.Elem()
	// }

	// if v.Kind() == reflect.Struct {
	// 	return fmt.Sprintf("%#v", arg)
	// }

	// return fmt.Sprint(arg)
	return commonerrors.FormatValue(arg)
}

func (l *logging) context2Attributes(ctx context.Context) []slog.Attr {
	args := []slog.Attr{}
	for _, key := range l.contextKeys {
		val, ok := ctx.Value(key).(string)
		if !ok {
			continue
		}
		args = append(args, slog.Any(key, val))
	}
	return args
}

func functionCallers() string {
	var builder strings.Builder
	builder.Grow(150)
	pc := make([]uintptr, 3)
	n := runtime.Callers(4, pc)
	if n == 0 {
		return "(no caller data)"
	}
	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)
	for {
		frame, more := frames.Next()
		name := frame.Func.Name()
		start := strings.LastIndex(name, "/")
		builder.WriteString("by ")
		builder.WriteString(name[start+1:])
		builder.WriteString(" at ")
		builder.WriteString(strconv.Itoa(frame.Line))
		builder.WriteString("\n")
		if !more {
			break
		}
	}

	return builder.String()
}
