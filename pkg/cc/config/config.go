package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/go-lib/certs"
	"github.com/aserto-dev/go-lib/logger"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Overrider is a func that mutates configuration
type Overrider func(*Config)

// Config holds the configuration for the app.
type Config struct {
	FileStoreRoot string        `mapstructure:"file_store_root"`
	DefaultDomain string        `mapstructure:"default_domain"`
	Logging       logger.Config `mapstructure:"logging"`

	Servers map[string]ServerCredentials `mapstructure:"servers"`
}

type ServerCredentials struct {
	Type     string `mapstructure:"type"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// Path is a string that points to a config file
type Path string

// NewConfig creates the configuration by reading env & files
func NewConfig(configPath Path, log *zerolog.Logger, overrides Overrider, certsGenerator *certs.Generator) (*Config, error) { // nolint // function will contain repeating statements for defaults
	configLogger := log.With().Str("component", "config").Logger()
	log = &configLogger

	cfg := new(Config)

	v := viper.New()

	file := string(configPath)
	if configPath != "" {
		exists, err := fileExists(file)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to determine if config file '%s' exists", configPath)
		}

		if exists {
			v.SetConfigType("yaml")
			v.AddConfigPath(".")
			v.SetConfigFile(file)
		}
	}

	v.SetEnvPrefix("POLICY")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// Set defaults
	v.SetDefault("file_store_root", filepath.Join(os.ExpandEnv("$HOME"), ".aserto", "policies-root"))
	v.SetDefault("default_domain", "registry.aserto.com") // policyregistry.io
	v.SetDefault("logging.log_level", "")
	v.SetDefault("logging.prod", false)

	configExists, err := fileExists(file)
	if err != nil {
		return nil, errors.Wrapf(err, "filesystem error")
	}

	if configExists {
		if err = v.ReadInConfig(); err != nil {
			return nil, errors.Wrapf(err, "failed to read config file '%s'", file)
		}
	}
	v.AutomaticEnv()

	err = v.UnmarshalExact(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config file")
	}

	if overrides != nil {
		overrides(cfg)
	}

	// This is where validation of config happens
	err = func() error {
		err = cfg.Logging.ParseLogLevel(zerolog.Disabled)
		if err != nil {
			return errors.Wrap(err, "failed to parse 'logging.log_level'")
		}

		return nil
	}()

	if err != nil {
		return nil, errors.Wrap(err, "failed to validate config file")
	}

	err = cfg.LoadCreds()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// NewLoggerConfig creates a new LoggerConfig
func NewLoggerConfig(configPath Path, overrides Overrider) (*logger.Config, error) {
	discardLogger := zerolog.New(ioutil.Discard)
	cfg, err := NewConfig(configPath, &discardLogger, overrides, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new config")
	}

	return &cfg.Logging, nil
}

func (c *Config) LoadCreds() error {
	path := os.ExpandEnv("$HOME/.aserto/policy-registries.yaml")

	if _, err := os.Stat(path); err == nil {
		contents, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read registry creds file [%s]", path)
		}

		err = yaml.Unmarshal(contents, &c.Servers)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal registry creds file [%s]", path)
		}

	} else if !os.IsNotExist(err) {
		return errors.Wrapf(err, "failed to determine if creds file [%s] exists", path)
	}

	return nil
}

func (c *Config) SaveCreds() error {
	path := os.ExpandEnv("$HOME/.aserto/policy-registries.yaml")

	cfgBytes, err := yaml.Marshal(c.Servers)
	if err != nil {
		return errors.Wrap(err, "failed to marshal registry creds for writing")
	}

	err = os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to create registry creds dir [%s]", filepath.Dir(path))
	}

	err = os.WriteFile(path, cfgBytes, 0600)
	if err != nil {
		return errors.Wrapf(err, "failed to write registry creds file [%s]", path)
	}

	return nil
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Wrapf(err, "failed to stat file '%s'", path)
	}
}

// func (c *Config) LoadCreds() error {
// 	path := os.ExpandEnv("$HOME/.aserto/policy-registries.yaml")

// 	if _, err := os.Stat(path); err == nil {
// 		contents, err := os.ReadFile(path)
// 		if err != nil {
// 			return errors.Wrapf(err, "failed to read registry creds file [%s]", path)
// 		}

// 		err = yaml.Unmarshal(contents, &c.Servers)
// 		if err != nil {
// 			return errors.Wrapf(err, "failed to unmarshal registry creds file [%s]", path)
// 		}

// 	} else if !os.IsNotExist(err) {
// 		return errors.Wrapf(err, "failed to determine if creds file [%s] exists", path)
// 	}

// 	for server, creds := range c.Servers {
// 		pass, err := keyring.Get(keyringPrefix+server, creds.Username)
// 		if err != nil {
// 			return err
// 		}
// 		creds.Password = pass
// 	}

// 	return nil
// }

// func (c *Config) SaveCreds() error {
// 	path := os.ExpandEnv("$HOME/.aserto/policy-registries.yaml")

// 	for server, creds := range c.Servers {
// 		err := keyring.Set(keyringPrefix+server, creds.Username, creds.Password)
// 		if err != nil {
// 			return err
// 		}
// 		creds.Password = ""
// 	}

// 	cfgBytes, err := yaml.Marshal(c.Servers)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to marshal registry creds for writing")
// 	}

// 	err = os.MkdirAll(filepath.Dir(path), 0700)
// 	if err != nil {
// 		return errors.Wrapf(err, "failed to create registry creds dir [%s]", filepath.Dir(path))
// 	}

// 	err = os.WriteFile(path, cfgBytes, 0600)
// 	if err != nil {
// 		return errors.Wrapf(err, "failed to write registry creds file [%s]", path)
// 	}

// 	return nil
// }
