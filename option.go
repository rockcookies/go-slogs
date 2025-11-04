package slogs

import "log/slog"

// Option configures a Logger.
//
// Options are passed to New or WithOptions to customize logger behavior.
type Option interface {
	apply(*Logger)
}

type optionFunc func(*Logger)

func (f optionFunc) apply(l *Logger) {
	f(l)
}

// WithCaller configures whether the logger should capture caller information.
//
// When enabled (default), the logger records the file and line number of the calling code.
// Disable this to improve performance when caller information is not needed.
//
// Example:
//
//	logger := slogs.New(handler, slogs.WithCaller(true))  // Enable caller info
//	logger := slogs.New(handler, slogs.WithCaller(false)) // Disable for performance
func WithCaller(enabled bool) Option {
	return optionFunc(func(l *Logger) {
		l.addCaller = enabled
	})
}

// WithCallerSkip adds the given number of stack frames to skip when capturing caller information.
//
// This is useful when wrapping the logger in your own logging functions.
// Each wrapper function should add WithCallerSkip(1) to ensure the correct caller is reported.
//
// Example:
//
//	// Wrapper function
//	func MyLog(logger *slogs.Logger, msg string) {
//		// Skip one additional frame to report the caller of MyLog, not MyLog itself
//		logger.WithOptions(slogs.WithCallerSkip(1)).Info(msg)
//	}
func WithCallerSkip(skip int) Option {
	return optionFunc(func(l *Logger) {
		l.callerSkip += skip
	})
}

// WithLevel sets the minimum log level for the logger.
//
// Log records below this level will be discarded. This is applied at the handler level,
// before any downstream handlers are called.
//
// Example:
//
//	logger := slogs.New(handler, slogs.WithLevel(slog.LevelInfo))
//	logger.Debug("This will not be logged") // Below LevelInfo
//	logger.Info("This will be logged")      // At or above LevelInfo
func WithLevel(level slog.Level) Option {
	return optionFunc(func(l *Logger) {
		l.handler = l.handler.WithLevel(level)
	})
}

// WithName adds a name to the logger's name chain.
//
// Names are displayed as a prefix in log messages, separated by dots for nested names.
// This helps identify the source of log messages in large applications.
//
// Empty names are ignored. Multiple WithName calls build up a chain of names.
//
// Example:
//
//	logger := slogs.New(handler, slogs.WithName("service"))
//	logger.Info("Started")  // Output: [service] Started
//
//	dbLogger := logger.WithOptions(slogs.WithName("database"))
//	dbLogger.Info("Connected")  // Output: [service.database] Connected
func WithName(name string) Option {
	return optionFunc(func(l *Logger) {
		if name == "" {
			return
		}

		// Clone handler to avoid modifying shared state
		l.handler = l.handler.Clone()
		l.handler.context.Names = append(l.handler.context.Names, name)
	})
}

// WithNameOverride replaces the entire logger name chain with a single name.
//
// Unlike WithName which appends to the name chain, this option clears any existing
// names and sets a new root name. Empty names are ignored.
//
// Example:
//
//	logger := slogs.New(handler, slogs.WithName("old"))
//	logger.Info("Message")  // Output: [old] Message
//
//	newLogger := logger.WithOptions(slogs.WithNameOverride("new"))
//	newLogger.Info("Message")  // Output: [new] Message
func WithNameOverride(name string) Option {
	return optionFunc(func(l *Logger) {
		if name == "" {
			return
		}

		// Clone handler to avoid modifying shared state
		l.handler = l.handler.Clone()
		l.handler.context.Names = []string{name}
	})
}
