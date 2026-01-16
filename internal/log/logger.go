package log

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var (
	logger         *log.Logger
	verboseEnabled = false
)

func init() {
	styles := log.DefaultStyles()
	styles.Levels[log.DebugLevel].
		Background(lipgloss.Color("63")).
		Foreground(lipgloss.Color("0"))

	//styles.Key = styles.Key.SetString("\n").PaddingLeft(2)
	//styles.Value = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))

	logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: false,
	})
	logger.SetStyles(styles)
	logger.SetLevel(log.InfoLevel)
}

func SetVerbose(v bool) {
	verboseEnabled = v
	if v {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}
}

func IsVerbose() bool {
	return verboseEnabled
}

func Debug(msg interface{}, keyvals ...interface{}) {
	logger.Debug(msg, keyvals...)
}
func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(msg interface{}, keyvals ...interface{}) {
	logger.Info(msg, keyvals...)
}

func Warn(msg interface{}, keyvals ...interface{}) {
	logger.Warn(msg, keyvals...)
}

func Error(msg interface{}, keyvals ...interface{}) {
	logger.Error(msg, keyvals...)
}

func Fatal(msg interface{}, keyvals ...interface{}) {
	logger.Fatal(msg, keyvals...)
}

func Print(msg interface{}, keyvals ...interface{}) {
	logger.Print(msg, keyvals...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}

func With(keyvals ...interface{}) *log.Logger {
	return logger.With(keyvals...)
}

func SetLevel(level log.Level) {
	logger.SetLevel(level)
}

func GetLevel() log.Level {
	return logger.GetLevel()
}

func SetPrefix(prefix string) {
	logger.SetPrefix(prefix)
}

func SetReportCaller(report bool) {
	logger.SetReportCaller(report)
}

func SetReportTimestamp(report bool) {
	logger.SetReportTimestamp(report)
}

func GetLogger() *log.Logger {
	return logger
}
