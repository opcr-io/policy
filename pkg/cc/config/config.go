package config

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aserto-dev/logger"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Overrider is a func that mutates configuration.
type Overrider func(*Config)

// Config holds the configuration for the app.
type Config struct {
	FileStoreRoot    string            `json:"file_store_root" yaml:"file_store_root"`
	DefaultDomain    string            `json:"default_domain" yaml:"default_domain"`
	Logging          logger.Config     `json:"logging" yaml:"logging"`
	CA               []string          `json:"ca" yaml:"ca"`
	Insecure         bool              `json:"insecure" yaml:"insecure"`
	TokenDefaults    map[string]string `json:"token_defaults" yaml:"token_defaults"`
	CredentialsStore credentials.Store `json:"-"`
}

// Path is a string that points to a config file.
type Path string

const (
	defaultDomain = "default-domain.cfg"
)

// NewConfig creates the configuration by reading env & files.
func NewConfig(configPath Path, log *zerolog.Logger, overrides Overrider) (*Config, error) { // nolint // function will contain repeating statements for defaults
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

	// Set defaults.
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to determine user home directory")
	}

	v.SetDefault("file_store_root", filepath.Join(home, ".policy"))
	v.SetDefault("logging.log_level", "")
	v.SetDefault("logging.prod", false)
	v.SetDefault("token_defaults", map[string]string{"ghcr.io": "TOKEN"})

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

	// This is where validation of config happens.
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

	cf, err := config.Load(cfg.FileStoreRoot)
	if err != nil {
		return nil, err
	}
	err = cfg.LoadDefaultDomain()
	if err != nil {
		log.Err(err).Msg("failed to load default-domain.cfg file")
	}
	cfg.CredentialsStore = cf.GetCredentialsStore(cfg.DefaultDomain)

	return cfg, nil
}

// NewLoggerConfig creates a new LoggerConfig.
func NewLoggerConfig(configPath Path, overrides Overrider) (*logger.Config, error) {
	discardLogger := zerolog.New(io.Discard)
	cfg, err := NewConfig(configPath, &discardLogger, overrides)
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
func (c *Config) SaveDefaultDomain() error {
	_, err := os.Stat(c.FileStoreRoot)
	if err != nil {
		err := os.Mkdir(c.FileStoreRoot, 0600)
		if err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(c.FileStoreRoot, defaultDomain), []byte(c.DefaultDomain), 0600)
}

func (c *Config) LoadDefaultDomain() error {
	defaultPath := filepath.Join(c.FileStoreRoot, defaultDomain)
	ok, err := fileExists(defaultPath)
	if err != nil {
		return err
	}
	if ok {
		domain, err := os.ReadFile(defaultPath)
		if err != nil {
			return err
		}
		c.DefaultDomain = string(domain)
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
