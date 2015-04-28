package ioc

type Logger interface {
	Infof(message string, args ...interface{})
}
