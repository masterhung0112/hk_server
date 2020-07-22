package mlog

import (
  "context"
  "go.uber.org/zap"
  "go.uber.org/zap/zapcore"

  "github.com/mattermost/logr"
)

type Field = zapcore.Field

var Int64 = zap.Int64
var Int32 = zap.Int32
var Int = zap.Int
var Uint32 = zap.Uint32
var String = zap.String
var Any = zap.Any
var Err = zap.Error
var NamedErr = zap.NamedError
var Bool = zap.Bool
var Duration = zap.Duration

type LoggerConfiguration struct {
  EnableConsole bool
  ConsoleJson   bool
  ConsoleLevel  string
  EnableFile    bool
  FileJson      bool
  FileLevel     string
  FileLocation  string
}

type Logger struct {
  zap          *zap.Logger
  consoleLevel zap.AtomicLevel
  fileLevel    zap.AtomicLevel
  logrLogger   *logr.Logger
}

func (l *Logger) Debug(message string, fields ...Field) {
  l.zap.Debug(message, fields...)
  if l.logrLogger != nil && isLevelEnabled(l.logrLogger, logr.Debug) {
    l.logrLogger.WithFields(zapToLogr(fields)).Debug(message)
  }
}

func (l *Logger) Info(message string, fields ...Field) {
  l.zap.Info(message, fields...)
  if l.logrLogger != nil && isLevelEnabled(l.logrLogger, logr.Info) {
    l.logrLogger.WithFields(zapToLogr(fields)).Info(message)
  }
}

func (l *Logger) Warn(message string, fields ...Field) {
  l.zap.Warn(message, fields...)
  if l.logrLogger != nil && isLevelEnabled(l.logrLogger, logr.Warn) {
    l.logrLogger.WithFields(zapToLogr(fields)).Warn(message)
  }
}

func (l *Logger) Error(message string, fields ...Field) {
  l.zap.Error(message, fields...)
  if l.logrLogger != nil && isLevelEnabled(l.logrLogger, logr.Error) {
    l.logrLogger.WithFields(zapToLogr(fields)).Error(message)
  }
}

func (l *Logger) Critical(message string, fields ...Field) {
  l.zap.Error(message, fields...)
  if l.logrLogger != nil && isLevelEnabled(l.logrLogger, logr.Error) {
    l.logrLogger.WithFields(zapToLogr(fields)).Error(message)
  }
}

func (l *Logger) Log(level LogLevel, message string, fields ...Field) {
  if l.logrLogger != nil && isLevelEnabled(l.logrLogger, logr.Level(level)) {
    l.logrLogger.WithFields(zapToLogr(fields)).Log(logr.Level(level), message)
  }
}

func (l *Logger) LogM(levels []LogLevel, message string, fields ...Field) {
  if l.logrLogger != nil {
    var logger *logr.Logger
    for _, lvl := range levels {
      if isLevelEnabled(l.logrLogger, logr.Level(lvl)) {
        // don't create logger with fields unless at least one level is active.
        if logger == nil {
          l := l.logrLogger.WithFields(zapToLogr(fields))
          logger = &l
        }
        logger.Log(logr.Level(lvl), message)
      }
    }
  }
}

func (l *Logger) Flush(cxt context.Context) error {
  if l.logrLogger != nil {
    return l.logrLogger.Logr().Flush() // TODO: use context when Logr lib supports it.
  }
  return nil
}

// ShutdownAdvancedLogging stops the logger from accepting new log records and tries to
// flush queues within the context timeout. Once complete all targets are shutdown
// and any resources released.
func (l *Logger) ShutdownAdvancedLogging(cxt context.Context) error {
  var err error
  if l.logrLogger != nil {
    err = l.logrLogger.Logr().Shutdown() // TODO: use context when Logr lib supports it.
    l.logrLogger = nil
  }
  return err
}

// ConfigAdvancedLoggingConfig (re)configures advanced logging based on the
// specified log targets. This is the easiest way to get the advanced logger
// configured via a config source such as file.
func (l *Logger) ConfigAdvancedLogging(targets LogTargetCfg) error {
  if l.logrLogger != nil {
    if err := l.ShutdownAdvancedLogging(context.Background()); err != nil {
      Error("error shutting down previous logger", Err(err))
    }
  }

  logr, err := newLogr(targets)
  l.logrLogger = logr
  return err
}

// AddTarget adds a logr.Target to the advanced logger. This is the preferred method
// to add custom targets or provide configuration that cannot be expressed via a
//config source.
func (l *Logger) AddTarget(target logr.Target) error {
  return l.logrLogger.Logr().AddTarget(target)
}
