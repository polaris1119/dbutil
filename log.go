package dbutil

type Logger interface {
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})

	Sqlf(format string, args ...interface{})
	Sqlln(args ...interface{})

	Infof(format string, args ...interface{})
	Infoln(args ...interface{})

	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
}
