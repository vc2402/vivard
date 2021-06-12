package dependencies

import "github.com/sirupsen/logrus"

type ConfigProvider interface {
	GetConfig(name string) interface{}
}

type Provider interface {
	Logger(name string) *logrus.Entry
	Config() ConfigProvider
}
