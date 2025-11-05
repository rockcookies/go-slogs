package slogs

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// RedirectStdLogAt redirects output from the standard library's package-global
// logger to the supplied logger at the specified level. Since slogs already
// handles caller annotations, timestamps, etc., it automatically disables the
// standard library's annotations and prefixing.
//
// It returns a function to restore the original prefix and flags and reset the
// standard library's output to os.Stderr.
func RedirectStdLogAt(logger *Logger, level slog.Level) (func(), error) {
	flags := log.Flags()
	prefix := log.Prefix()

	handler := logger.Handler()
	slog.SetDefault(slog.New(handler))

	capturePC := log.Flags()&(log.Lshortfile|log.Llongfile) != 0
	log.SetFlags(0) // we want just the log message, no time or location
	log.SetPrefix("")
	log.SetOutput(&handlerWriter{handler, &level, capturePC})

	return func() {
		log.SetFlags(flags)
		log.SetPrefix(prefix)
		log.SetOutput(os.Stderr)
	}, nil
}

// handlerWriter is an io.Writer that calls a Handler.
// It is used to link the default log.Logger to the default slogs.Logger.
type handlerWriter struct {
	h         slog.Handler
	level     slog.Leveler
	capturePC bool
}

func (w *handlerWriter) Write(buf []byte) (int, error) {
	level := w.level.Level()
	if !w.h.Enabled(context.Background(), level) {
		return 0, nil
	}
	var pc uintptr
	if !w.capturePC {
		// skip [runtime.Callers, w.Write, Logger.Output, log.Print]
		var pcs [1]uintptr
		runtime.Callers(4, pcs[:])
		pc = pcs[0]
	}

	// Remove final newline.
	origLen := len(buf) // Report that the entire buf was written.
	buf = bytes.TrimSuffix(buf, []byte{'\n'})
	r := slog.NewRecord(time.Now(), level, string(buf), pc)
	return origLen, w.h.Handle(context.Background(), r)
}
