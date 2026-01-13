package config

import (
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	cfg *Config
)

type option struct {
	configFolder string
	configFile   string
	configType   string
}

type Option func(*option)

func Init(opts ...Option) error {
	opt := &option{
		configFolder: "./configs/",
		configFile:   "config",
		configType:   "yaml",
	}

	for _, optFunc := range opts {
		optFunc(opt)
	}

	envPath := opt.configFolder + ".env"

	if err := godotenv.Load(envPath); err != nil {
		slog.Warn("Warning: .env file not found, relying on system env vars")
	}

	viper.AddConfigPath(opt.configFolder)
	viper.SetConfigName(opt.configFile)
	viper.SetConfigType(opt.configType)

	viper.AutomaticEnv()

	setDefault()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	cfg = new(Config)
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}

	return cfg.postprocess()
}

func WithConfigFolder(path string) Option {
	return func(o *option) { o.configFolder = path }
}

func WithConfigFile(name string) Option {
	return func(o *option) { o.configFile = name }
}

func WithConfigType(configType string) Option {
	return func(o *option) {
		o.configType = configType
	}
}

func Get() *Config {
	if cfg == nil {
		return &Config{}
	}
	return cfg
}

func setDefault() {
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("ENV", "development")
	viper.SetDefault("GLOBAL_TIMEOUT", 5)

	viper.SetDefault("GARUDA_PATH", "")
	viper.SetDefault("LION_PATH", "")
	viper.SetDefault("AIRASIA_PATH", "")
	viper.SetDefault("BATIK_PATH", "")
	viper.SetDefault("AGGREGATOR_TIMEOUT", 5)
}

func (c *Config) postprocess() error {
	return nil
}
