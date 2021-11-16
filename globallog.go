package elog

type ILog interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})

	Warn(v ...interface{})
	Warnf(format string, v ...interface{})

	Error(v ...interface{})
	Errorf(format string, v ...interface{})
}

var GlobalLog ILog

func Debug(v ...interface{}) {
	GlobalLog.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	GlobalLog.Debugf(format, v...)
}

func Info(v ...interface{}) {
	GlobalLog.Info(v...)
}

func Infof(format string, v ...interface{}) {
	GlobalLog.Infof(format, v...)
}

func Warn(v ...interface{}) {
	GlobalLog.Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	GlobalLog.Warnf(format, v...)
}

func Error(v ...interface{}) {
	GlobalLog.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	GlobalLog.Errorf(format, v...)
}
