package config

import (
	"strings"

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

	viper.AddConfigPath(opt.configFolder)
	viper.SetConfigName(opt.configFile)
	viper.SetConfigType(opt.configType)

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
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

	return nil
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

	// Default paths (relative to root)
	viper.SetDefault("PROVIDERS.GARUDA_PATH", "mock_data/garuda.json")
	viper.SetDefault("PROVIDERS.LION_PATH", "mock_data/lion.json")
	viper.SetDefault("PROVIDERS.AIRASIA_PATH", "mock_data/airasia.json")
	viper.SetDefault("PROVIDERS.BATIK_PATH", "mock_data/batik.json")
}
