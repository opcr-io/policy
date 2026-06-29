package runtime

import (
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/keys"
	bundleplugin "github.com/open-policy-agent/opa/v1/plugins/bundle"
	"github.com/open-policy-agent/opa/v1/plugins/logs"
	"github.com/open-policy-agent/opa/v1/plugins/status"
	"github.com/open-policy-agent/opa/v1/topdown/cache"
)

type Config struct {
	LocalBundles                  LocalBundlesConfig `json:"local_bundles"`
	InstanceID                    string             `json:"instance_id"`
	PluginsErrorLimit             int                `json:"plugins_error_limit"`
	GracefulShutdownPeriodSeconds int                `json:"graceful_shutdown_period_seconds"`
	MaxPluginWaitTimeSeconds      int                `json:"max_plugin_wait_time_seconds"`
	Flags                         Flags              `json:"flags"`
	Config                        OPAConfig          `json:"config"`
}

type Flags struct {
	EnableStatusPlugin bool `json:"enable_status_plugin"`
}

type LocalBundlesConfig struct {
	Watch              bool                       `json:"watch"`
	LocalPolicyImage   string                     `json:"local_policy_image"`
	FileStoreRoot      string                     `json:"file_store_root"`
	Paths              []string                   `json:"paths"`
	Ignore             []string                   `json:"ignore"`
	SkipVerification   bool                       `json:"skip_verification"`
	VerificationConfig *bundle.VerificationConfig `json:"verification_config"`
}

type OPAConfig struct {
	Services                     map[string]any                  `json:"services,omitempty"`
	Labels                       map[string]string               `json:"labels,omitempty"`
	Bundles                      map[string]*bundleplugin.Source `json:"bundles,omitempty"`
	DecisionLogs                 *logs.Config                    `json:"decision_logs,omitempty"`
	Status                       *status.Config                  `json:"status,omitempty"`
	Plugins                      map[string]any                  `json:"plugins,omitempty"`
	Keys                         map[string]*keys.Config         `json:"keys,omitempty"`
	DefaultDecision              *string                         `json:"default_decision,omitempty"`
	DefaultAuthorizationDecision *string                         `json:"default_authorization_decision,omitempty"`
	Caching                      *cache.Config                   `json:"caching,omitempty"`
	PersistenceDirectory         *string                         `json:"persistence_directory,omitempty"`
}
