package dependencies

import (
	"go.uber.org/zap"
)

type ConfigProvider interface {
	GetConfig(name string) interface{}
}

type Provider interface {
	Logger(name string) *zap.Logger
	Config() ConfigProvider
}
