package logging

import (
	"io"
	"log/slog"
	"os"
	"path"
	"time"

	logi "gitcode.yealink.com/server/devops/yops/ykit/log"
	timex "gitcode.yealink.com/server/devops/yops/ykit/util/time"
	"github.com/docker/go-units"
	"github.com/natefinch/lumberjack"
)

type LogConfig struct {
	Format    string         `json:"format" yaml:"format" default:"json" usage:"log format (json|console)"`
	Level     string         `json:"level" yaml:"level" default:"info" usage:"log level (debug|info|warn|error)"`
	Verbosity int            `json:"verbosity" yaml:"verbosity" default:"3" usage:"log verbosity (0-4)"`
	Attrs     map[string]any `json:"attrs" yaml:"attrs" default:"{}" usage:"log attrs"`
	Output    struct {
		Stdout struct {
			Enable bool `json:"enable" yaml:"enable" default:"false" usage:"enable stdout log"`
		} `json:"stdout" yaml:"stdout"`
		Stderr struct {
			Enable bool `json:"enable" yaml:"enable" default:"false" usage:"enable stderr log"`
		} `json:"stderr" yaml:"stderr"`
		File struct {
			Enable     bool   `json:"enable" yaml:"enable" default:"false" usage:"enable file log"`
			Path       string `json:"path" yaml:"path" default:"/var/log/yauth/yauth.log" usage:"log file path"`
			MaxSize    string `json:"maxSize" yaml:"maxSize" default:"10mb" usage:"log file max size using SI(decimal) standard (K|mb|Gb...)"`
			MaxAge     string `json:"maxAge" yaml:"maxAge" default:"7d" usage:"log file max age (d|h|m|s)"`
			MaxBackups int    `json:"maxBackups" yaml:"maxBackups" default:"10" usage:"log file max backups"`
			Compress   bool   `json:"compress" yaml:"compress" default:"true" usage:"enable log file compress"`
		} `json:"file" yaml:"file"`
	} `json:"output" yaml:"output"`
}

func NewLogHandler(config LogConfig) slog.Handler {
	writers := make([]io.Writer, 0)
	if config.Output.Stdout.Enable {
		writers = append(writers, os.Stdout)
	}
	if config.Output.Stderr.Enable {
		writers = append(writers, os.Stderr)
	}
	if config.Output.File.Enable {
		maxSize, err := units.FromHumanSize(config.Output.File.MaxSize)
		if err != nil {
			panic(err)
		}
		maxAgeDur, err := timex.ParseDuration(config.Output.File.MaxAge)
		if err != nil {
			panic(err)
		}
		err = os.MkdirAll(path.Dir(config.Output.File.Path), 0o755)
		if err != nil {
			panic(err)
		}
		writers = append(writers, &lumberjack.Logger{
			Filename:   config.Output.File.Path,
			MaxSize:    int(maxSize / units.MB),
			MaxAge:     int(maxAgeDur / (24 * time.Hour)),
			MaxBackups: config.Output.File.MaxBackups,
			LocalTime:  false,
			Compress:   config.Output.File.Compress,
		})
	}

	return logi.NewHandler(
		logi.HandlerOptions{
			Format:         config.Format,
			Level:          config.Level,
			Attrs:          config.Attrs,
			Writers:        writers,
			Verbosity:      config.Verbosity,
			VerbosityHooks: []logi.VerbosityHook{},
			CallerSkip:     11,
			ReplaceAttrs:   []logi.ReplaceAttrFunc{},
		},
	)
}

func NewLogHandlers(configs ...LogConfig) []slog.Handler {
	handlers := make([]slog.Handler, 0)
	for _, config := range configs {
		handlers = append(handlers, NewLogHandler(config))
	}
	return handlers
}

func NewFanoutLogHandler(configs ...LogConfig) slog.Handler {
	return logi.NewFanOutHandler(NewLogHandlers(configs...)...)
}
