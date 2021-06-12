package vivard

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

const (
	sConsole               = "console"
	sLevel                 = "level"
	sFormatter             = "formatter"
	sHooks                 = "hooks"
	sMain                  = "main"
	sDateFormat            = "dateformat"
	sFormatterType         = "type"
	sFormatterTypeJSON     = "json"
	sFormatterTypeText     = "text"
	sFormatterShrinkFields = "shrinkfields"
	sHookType              = "type"
	sHookTypeFile          = "file"
	sPath                  = "path"
	sLevels                = "levels"
	sRotating              = "rotating"
	sRotatingRotationTime  = "rotationtime"
	sRotatingMaxAge        = "maxage"
)

type rotatingFileConfig struct {
	maxAge       time.Duration
	rotationTime time.Duration
}
type logFileConfig struct {
	path      string
	rotating  *rotatingFileConfig
	levels    []logrus.Level
	formatter logrus.Formatter
}

func initLogrus(config map[string]interface{}) (*logrus.Logger, error) {
	var logger *logrus.Logger = logrus.StandardLogger()
	if config != nil && len(config) > 0 {
		console, _ := config[sConsole].(bool)
		lev, ok := config[sLevel].(string)
		if !ok {
			lev = "debug"
		}
		level, err := logrus.ParseLevel(lev)
		if err != nil {
			return nil, err
		}
		var formatter logrus.Formatter
		formCfg, _ := config[sFormatter].(map[string]interface{})
		if formCfg != nil {
			formatter, err = initLogReadFormatterConfig(formCfg)
			if err != nil {
				return nil, fmt.Errorf("problem while reading formatter config: %w", err)
			}
		} else {
			formatter = &logrus.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000"}
		}
		var main *logFileConfig
		hooks := map[string]*logFileConfig{}
		hh, ok := config[sHooks].(map[string]interface{})
		if ok {
			for n, h := range hh {
				hc, ok := h.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("hook description should be an object: %s", n)
				}
				hook, err := initLogReadFileConfig(hc)
				if err != nil {
					return nil, fmt.Errorf("problem while reading hook's '%s' config: %w", n, err)
				}
				if n == sMain && !console {
					main = hook
				}
				hooks[n] = hook

			}

		}
		logger.SetLevel(level)
		logger.SetFormatter(formatter)
		if main != nil {
			if main.rotating == nil {
				main.rotating = &rotatingFileConfig{maxAge: -1, rotationTime: -1}
			}
			w, err := main.rotating.getWriter(main.path)
			if err != nil {
				return nil, fmt.Errorf("problem while creating writer for main hook: %w", err)
			}
			logger.SetOutput(w)
		}

		for _, lc := range hooks {
			f := formatter
			if lc.formatter != nil {
				f = lc.formatter
			}
			if lc.rotating != nil {
				writer, err := lc.rotating.getWriter(lc.path)
				if err != nil {
					return nil, fmt.Errorf("problem while creating writer: %w", err)
				}
				writerMap := lfshook.WriterMap{}
				for _, l := range lc.levels {
					writerMap[l] = writer
				}
				logger.AddHook(lfshook.NewHook(writerMap, f))
			} else {
				pathMap := lfshook.PathMap{}
				for _, l := range lc.levels {
					pathMap[l] = lc.path
				}
				logger.AddHook(lfshook.NewHook(pathMap, f))
			}
		}
	}
	return logger, nil
	// logrus_mate.Hijack(
	// 	logrus.StandardLogger(),
	// 	logrus_mate.ConfigFile(fileName),
	// )
}

func initLogReadFileConfig(cfg map[string]interface{}) (*logFileConfig, error) {
	var ret *logFileConfig
	tip, ok := cfg[sHookType].(string)
	if !ok {
		tip = sHookTypeFile
	}
	var err error
	var formatter logrus.Formatter
	if form, ok := cfg[sFormatter].(map[string]interface{}); ok {
		formatter, err = initLogReadFormatterConfig(form)
		if err != nil {
			return nil, fmt.Errorf("invalid formatter config: %w", err)
		}
	}
	var levels []logrus.Level
	if ls, ok := cfg[sLevels].([]interface{}); ok {
		levels = make([]logrus.Level, len(ls))
		for i, l := range ls {
			if lev, ok := l.(string); ok {
				level, err := logrus.ParseLevel(lev)
				if err != nil {
					return nil, fmt.Errorf("can't recognise level %s", lev)
				}
				levels[i] = level
			} else {
				return nil, fmt.Errorf("can't recognise level %v", l)
			}
		}
	} else {
		levels = []logrus.Level{
			logrus.TraceLevel,
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
			logrus.FatalLevel,
		}
	}

	switch tip {
	case sHookTypeFile:
		path, ok := cfg[sPath].(string)
		if !ok {
			return nil, errors.New("path should be set for file hook")
		}
		ret = &logFileConfig{path: path, formatter: formatter, levels: levels}
		if rot, ok := cfg[sRotating]; ok {
			ret.rotating = &rotatingFileConfig{}
			if rotCfg, ok := rot.(map[string]interface{}); ok {
				if rotTime, ok := rotCfg[sRotatingRotationTime].(string); ok {
					rt, err := time.ParseDuration(rotTime)
					if err != nil {
						return nil, fmt.Errorf("while parsing rotation time duration: %w", err)
					}
					ret.rotating.rotationTime = rt
				}
				if ageCfg, ok := rotCfg[sRotatingRotationTime].(string); ok {
					age, err := time.ParseDuration(ageCfg)
					if err != nil {
						return nil, fmt.Errorf("while parsing rotation age: %w", err)
					}
					ret.rotating.maxAge = age
				}
			}
		}
	default:
		err = fmt.Errorf("undefined hook type: %s", tip)
	}
	return ret, err
}

func initLogReadFormatterConfig(cfg map[string]interface{}) (logrus.Formatter, error) {

	timestampFormat := "2006-01-02 15:04:05.000"
	disableTimestamp := false
	if tf, ok := cfg[sDateFormat]; ok {
		switch v := tf.(type) {
		case bool:
			if !v {
				disableTimestamp = true
			}
		case string:
			if v == "" || v == "off" {
				disableTimestamp = true
			} else {
				timestampFormat = v
			}
		}
	}
	var fieldMap logrus.FieldMap
	if sn, ok := cfg[sFormatterShrinkFields].(bool); ok && sn {
		fieldMap = logrus.FieldMap{
			logrus.FieldKeyTime:  "t",
			logrus.FieldKeyLevel: "l",
			logrus.FieldKeyMsg:   "m",
		}
	}

	var ret logrus.Formatter

	if ft, ok := cfg[sFormatterType].(string); ok {
		switch ft {
		case sFormatterTypeJSON:
			ret = &logrus.JSONFormatter{
				TimestampFormat:  timestampFormat,
				DisableTimestamp: disableTimestamp,
				FieldMap:         fieldMap,
			}
		case sFormatterTypeText:
			ret = &logrus.TextFormatter{
				TimestampFormat:  timestampFormat,
				DisableTimestamp: disableTimestamp,
				// PadLevelText: true,
				FieldMap: fieldMap,
			}
		default:
			return nil, fmt.Errorf("unknown formatter type: %s", ft)
		}
	}
	return ret, nil
}

func (rfc *rotatingFileConfig) getWriter(path string) (io.Writer, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
		logrus.WithFields(
			logrus.Fields{
				"path": path,
				"err":  err,
			}).Warn("can't get abs path")
	}
	return rotatelogs.New(
		absPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(absPath),
		rotatelogs.WithMaxAge(rfc.maxAge),
		rotatelogs.WithRotationTime(rfc.rotationTime),
	)
}
