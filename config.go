package vivard

import (
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type ConfigProvider interface {
	GetConfigValue(key string) interface{}
}

type ViperConfig struct {
	vip *viper.Viper
}

func NewViperConfig() ViperConfig {
	return ViperConfig{viper.GetViper()}
}

func NewViperConfigForViper(v *viper.Viper) ViperConfig {
	return ViperConfig{v}
}

func (vc ViperConfig) GetConfigValue(key string) interface{} {
	return vc.vip.Get(strings.ToLower(key))
}

type configProvider struct {
	provider ConfigProvider
	next     *configProvider
	priority int
}

func (eng *Engine) RegisterConfigProvider(p ConfigProvider, priority int) {
	cp := eng.config
	var prev *configProvider
	for cp != nil && cp.priority > priority {
		prev = cp
		cp = cp.next
	}
	newProv := &configProvider{provider: p, priority: priority, next: cp}
	if prev != nil {
		prev.next = newProv
	} else {
		eng.config = newProv
	}
}

func (eng *Engine) ConfValue(key string) interface{} {
	var ret interface{}
	for cp := eng.config; cp != nil; cp = cp.next {
		if ret = cp.provider.GetConfigValue(key); ret != nil {
			return ret
		}
	}
	return ret
}

func (eng *Engine) ConfString(key string, def ...string) string {
	if str, ok := eng.ConfValue(key).(string); ok {
		return str
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func (eng *Engine) ConfInt(key string, def ...int) int {
	if val := eng.ConfValue(key); val != nil {
		switch v := val.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			ret, _ := strconv.Atoi(v)
			return ret

		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

func (eng *Engine) ConfInt64(key string, def ...int64) int64 {
	if val := eng.ConfValue(key); val != nil {
		switch v := val.(type) {
		case int:
			return int64(v)
		case int32:
			return int64(v)
		case int64:
			return v
		case float64:
			return int64(v)
		case string:
			ret, _ := strconv.ParseInt(v, 10, 64)
			return ret

		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
