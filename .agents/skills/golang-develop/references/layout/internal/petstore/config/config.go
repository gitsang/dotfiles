package config

import (
	"golang-develop-example/pkg/logging"

	"gitcode.yealink.com/server/devops/yops/ykit/configer"
	"github.com/spf13/cobra"
)

type Config struct {
	Timezone string `default:"Asia/Shanghai"`

	Log logging.LogHandlerSetConfig

	Server struct {
		Http struct {
			Network   string `default:"tcp"`
			Addr      string `default:":6"`
			Timeout   string `default:"3s"`
			ApiPrefix string `default:"/petstore/api"`
		}

		Grpc struct {
			Network string `default:"tcp"`
			Addr    string `default:":2885"`
		}

		Passkey struct {
			RPID      string
			RPOrigins []string
		}
	}

	Data struct {
		Database struct {
			Driver string `default:"mysql"`
			MySQL  struct {
				DSN string `default:"root:password@tcp(127.0.0.1:3306)/app1?charset=utf8mb4&parseTime=True&loc=Local"`
			}
			Sqlite struct {
				Path string
			}
			Postgres struct {
				DSN string `default:"root:password@tcp(127.0.0.1:3306)/app1?charset=utf8mb4&parseTime=True&loc=Local"`
			}
		}

		Cache struct {
			Driver string `default:"memory"`
			Redis  struct {
				Addresses  string   `default:"127.0.0.1:6379"`
				Username   string   `default:""`
				Password   string   `default:""`
				MasterName string   `default:""`
				Db         int      `default:"0"`
				Namespaces []string `default:"app1"`
			}
			Memory struct {
				Namespaces []string `default:"app1"`
			}
		}

		Mq struct {
			Driver string `default:"memory"`
			Redis  struct {
				Addresses  string   `default:"127.0.0.1:6379"`
				Username   string   `default:""`
				Password   string   `default:""`
				MasterName string   `default:""`
				Db         int      `default:"0"`
				Namespaces []string `default:"app1"`
			}
			Memory struct {
				Namespaces []string `default:"app1"`
			}
			KeepaliveInterval string `default:"5s"`
			ConsumeTimeout    string `default:"15s"`
			MaxRetry          int64  `default:"3"`
		}
	}
}

func NewConfiger(rootCmd *cobra.Command) *configer.Configer {
	return configer.New(
		configer.WithTemplate((*Config)(nil)),
		configer.WithEnvBind(
			configer.WithEnvPrefix("PETSTORE"),
			configer.WithEnvDelim("_"),
		),
		configer.WithFlagBind(
			configer.WithCommand(rootCmd),
			configer.WithFlagPrefix("config"),
			configer.WithFlagDelim("."),
		),
	)
}
