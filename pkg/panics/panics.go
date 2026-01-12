package panics

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/breaker"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/grpc"
)

var (
	env           string
	filepath      string
	tagString     string
	customMessage string
	cb            *breaker.Breaker
	ErrorPanic    = errors.New("panic happened")
)

type Tags map[string]string

type Options struct {
	Env           string
	Filepath      string
	SentryDSN     string
	Tags          Tags
	CustomMessage string
	DontLetMeDie  bool
}

func SetOptions(o *Options) {
	filepath = o.Filepath

	env = o.Env

	var tmp []string
	for key, val := range o.Tags {
		tmp = append(tmp, fmt.Sprintf("`%s: %s`", key, val))
	}
	tagString = strings.Join(tmp, " | ")

	customMessage = o.CustomMessage

	// set circuit breaker to nil
	if o.DontLetMeDie {
		cb = nil
	}
}

func init() {
	env = os.Getenv("GO_ENV")
	// circuitbreaker to let apps died when got too many panics
	cb = breaker.New(3, 2, time.Minute*1)
}

// CaptureHandler handle panic on http handler.
func CaptureHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, _ := httputil.DumpRequest(r, true)
		defer func() {
			if !recoveryBreak() {
				r := panicRecover(recover())
				if r != nil {
					publishError(r, request, true)
					http.Error(w, r.Error(), http.StatusInternalServerError)
				}
			}
		}()
		h.ServeHTTP(w, r)
	}
}

// CaptureHTTPRouterHandler handle panic on httprouter handler.
func CaptureHTTPRouterHandler(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		request, _ := httputil.DumpRequest(r, true)
		defer func() {
			if !recoveryBreak() {
				r := panicRecover(recover())
				if r != nil {
					publishError(r, request, true)
					http.Error(w, r.Error(), http.StatusInternalServerError)
				}
			}
		}()
		h(w, r, ps)
	}
}

// CaptureNegroniHandler handle panic on negroni handler.
func CaptureNegroniHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	request, _ := httputil.DumpRequest(r, true)
	defer func() {
		if !recoveryBreak() {
			r := panicRecover(recover())
			if r != nil {
				publishError(r, request, true)
				http.Error(w, r.Error(), http.StatusInternalServerError)
			}
		}
	}()
	next(w, r)
}

// Capture will publish any errors
func Capture(err string, message ...string) {
	var tmp string
	for i, val := range message {
		if i == 0 {
			tmp += val
		} else {
			tmp += fmt.Sprintf("\n\n%s", val)
		}
	}

	publishError(errors.New(err), []byte(tmp), false)
}

// Capture will publish any errors with stack trace
func CaptureWithStackTrace(err string, message ...string) {
	var tmp string
	for i, val := range message {
		if i == 0 {
			tmp += val
		} else {
			tmp += fmt.Sprintf("\n\n%s", val)
		}
	}

	publishError(errors.New(err), []byte(tmp), true)
}

// HTTPRecoveryMiddleware act as middleware that capture panics standard in http handler
func HTTPRecoveryMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		request, _ := httputil.DumpRequest(r, true)
		defer func() {
			if !recoveryBreak() {
				rcv := panicRecover(recover())
				if rcv != nil {
					// log the panic
					fmt.Fprintf(os.Stderr, "Panic: %+v Request-Id:%s\n", rcv, r.Header.Get("X-Request-Id"))
					debug.PrintStack()
					publishError(rcv, request, true)
					http.Error(w, rcv.Error(), http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// UnaryServerInterceptor intercept the execution of a unary RPC on the server when panic happen
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if !recoveryBreak() {
			rcv := panicRecover(recover())
			if rcv != nil {
				fmt.Fprintf(os.Stderr, "Panic: %+v\n", rcv)
				debug.PrintStack()
				publishError(rcv, nil, true)
			}
		}
	}()
	return handler(ctx, req)
}

func panicRecover(rc interface{}) error {
	if cb != nil {
		r := cb.Run(func() error {
			return recovery(rc)
		})
		return r
	}
	return recovery(rc)
}

func recoveryBreak() bool {
	if cb == nil {
		return false
	}

	if err := cb.Run(func() error {
		return nil
	}); err == breaker.ErrBreakerOpen {
		return true
	}
	return false
}

func recovery(r interface{}) error {
	var err error
	if r != nil {
		switch t := r.(type) {
		case string:
			err = errors.New(t)
		case error:
			err = t
		default:
			err = errors.New("unknown error")
		}
	}
	return err
}

func publishError(errs error, reqBody []byte, withStackTrace bool) {
	var text string
	var snip string
	var buffer bytes.Buffer
	errorStack := debug.Stack()
	buffer.WriteString(fmt.Sprintf(`[%s] *%s*`, env, errs.Error()))

	if len(tagString) > 0 {
		buffer.WriteString(" | " + tagString)
	}

	if customMessage != "" {
		buffer.WriteString("\n" + customMessage + "\n")
	}

	if reqBody != nil {
		buffer.WriteString(fmt.Sprintf(" ```%s``` ", string(reqBody)))
	}
	text = buffer.String()

	if errorStack != nil && withStackTrace {
		snip = fmt.Sprintf("```\n%s```", string(errorStack))
	}
	slog.Info(fmt.Sprintf("[panics] %s %s", text, snip))
	if filepath != "" {
		go func() {
			fp := fmt.Sprintf("%s/panics.log", filepath)
			file, err := os.OpenFile(fp, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err != nil {
				slog.Info(fmt.Sprintf("[panics] failed to open file %s", fp))
				return
			}
			file.Write([]byte(text))
			file.Write([]byte(snip + "\r\n"))
			file.Close()
		}()
	}
}
