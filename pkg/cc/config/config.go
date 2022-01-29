package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/go-utils/certs"
	"github.com/aserto-dev/go-utils/logger"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// Overrider is a func that mutates configuration
type Overrider func(*Config)

// Config holds the configuration for the app.
type Config struct {
	FileStoreRoot string        `json:"file_store_root" yaml:"file_store_root"`
	DefaultDomain string        `json:"default_domain" yaml:"default_domain"`
	Logging       logger.Config `json:"logging" yaml:"logging"`
	CA            []string      `json:"ca" yaml:"ca"`
	Insecure      bool          `json:"insecure" yaml:"insecure"`

	Servers map[string]ServerCredentials `json:"-" yaml:"-"`
}

type ServerCredentials struct {
	Type     string `json:"type" yaml:"type"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Default  bool   `json:"default" yaml:"default"`
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
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to determine user home directory")
	}
	v.SetDefault("file_store_root", filepath.Join(home, ".policy"))
	v.SetDefault("default_domain", "opcr.io")
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

	err = v.UnmarshalExact(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "json"
	})
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
	if cfg.DefaultDomain == "" || cfg.DefaultDomain == "opcr.io" {
		for key, val := range cfg.Servers {
			if val.Default {
				cfg.DefaultDomain = key
				break
			}
		}
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

func (c *Config) PoliciesRoot() string {
	return filepath.Join(c.FileStoreRoot, "policies-root")
}

func (c *Config) ReplHistoryFile() string {
	return filepath.Join(c.FileStoreRoot, "repl_history")
}

func (c *Config) LoadCreds() error {
	path := filepath.Join(c.FileStoreRoot, "policy-registries.yaml")

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
	path := filepath.Join(c.FileStoreRoot, "policy-registries.yaml")

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
