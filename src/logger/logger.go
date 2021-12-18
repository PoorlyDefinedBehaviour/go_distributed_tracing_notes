package logger

import (
	"runtime"
	"strings"
	"time"

	zerologpkg "github.com/rs/zerolog"
	zerolog "github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

func init() {
	zerologpkg.TimeFieldFormat = zerologpkg.TimeFormatUnix
	zerologpkg.ErrorStackMarshaler = pkgerrors.MarshalStack
}

type Tracer func()

type T interface {
	Trace(string) Tracer
	Info(string)
	Error(string)
	Debug(string)
	Prefix(string)
}

type Event struct {
	name string
}

var Events = struct {
	TraceStart Event
	TraceEnd   Event
	Info       Event
	Error      Event
	Debug      Event
}{
	TraceStart: Event{"trace_start"},
	TraceEnd:   Event{"trace_end"},
	Info:       Event{"info"},
	Error:      Event{"error"},
	Debug:      Event{"debug"},
}

type EventHandler func(event Event, message string)

type logger struct {
	prefix     string
	beforeEach []EventHandler
}

func New() logger {
	return logger{}
}

func caller(skip int) string {
	// get caller function path

	pc := make([]uintptr, 15)
	runtime.Callers(skip, pc)
	n := runtime.Callers(skip, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	path := strings.Split(frame.Function, "/")

	return path[len(path)-1]
}

func (log *logger) BeforeEach(handler EventHandler) {
	log.beforeEach = append(log.beforeEach, handler)
}

func (log *logger) callBeforeEachHandlers(event Event, message string) {
	for _, handler := range log.beforeEach {
		handler(event, message)
	}
}

func (log *logger) Trace(message string) Tracer {
	log.callBeforeEachHandlers(Events.TraceStart, message)

	start := time.Now()

	zerolog.Info().
		Caller(1).
		Str("correlation_id", log.prefix).
		Msgf("[END - %d(%dms)] %s", time.Now().UnixMilli(), time.Since(start).Milliseconds(), message)

	return func() {
		log.callBeforeEachHandlers(Events.TraceEnd, message)

		zerolog.Info().
			Caller(1).
			Str("correlation_id", log.prefix).
			Msgf("[END - %d(%dms)] %s", time.Now().UnixMilli(), time.Since(start).Milliseconds(), message)
	}
}

func (log *logger) Info(message string) {
	log.callBeforeEachHandlers(Events.Info, message)

	zerolog.Info().Caller(1).Str("correlation_id", log.prefix).Msg(message)
}

func (log *logger) Error(message string) {
	log.callBeforeEachHandlers(Events.Error, message)

	zerolog.Error().Stack().Str("correlation_id", log.prefix).Msg(message)
}

func (log *logger) Debug(message string) {
	log.callBeforeEachHandlers(Events.Debug, message)

	zerolog.Debug().Caller(1).Str("correlation_id", log.prefix).Msg(message)
}

func (log *logger) Prefix(prefix string) {
	log.prefix = prefix
}
