package slogs

import "log/slog"

type Option interface {
	apply(*Logger)
}

type optionFunc func(*Logger)

func (f optionFunc) apply(l *Logger) {
	f(l)
}

func WithCaller(enabled bool) Option {
	return optionFunc(func(l *Logger) {
		l.addCaller = enabled
	})
}

func WithCallerSkip(skip int) Option {
	return optionFunc(func(l *Logger) {
		l.callerSkip += skip
	})
}

func WithLevel(level slog.Level) Option {
	return optionFunc(func(l *Logger) {
		l.handler = l.handler.WithLevel(level)
	})
}

func WithName(name string) Option {
	return optionFunc(func(l *Logger) {
		if name == "" {
			return
		}

		l.handler.context.Names = append(l.handler.context.Names, name)
	})
}

func WithNameOverride(name string) Option {
	return optionFunc(func(l *Logger) {
		if name == "" {
			return
		}

		l.handler.context.Names = []string{name}
	})
}
