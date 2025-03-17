module github.com/opcr-io/policy

go 1.24

toolchain go1.24.6

require (
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/alecthomas/kong v1.12.1
	github.com/aserto-dev/clui v0.8.3
	github.com/aserto-dev/logger v0.0.9
	github.com/aserto-dev/runtime v1.6.1-0.20250805112623-bcdae0d8d240
	github.com/aserto-dev/scc-lib v0.1.4
	github.com/containerd/containerd v1.7.25
	github.com/containerd/errdefs v1.0.0
	github.com/distribution/reference v0.6.0
	github.com/docker/cli v27.4.1+incompatible
	github.com/dustin/go-humanize v1.0.1
	github.com/go-viper/mapstructure/v2 v2.4.0
	github.com/google/wire v0.6.0
	github.com/magefile/mage v1.15.0
	github.com/opcr-io/oras-go/v2 v2.0.0-20231122155130-eb4260d8a0ae
	github.com/opcr-io/policy/oci v0.2.16
	github.com/opcr-io/policy/parser v0.2.16
	github.com/open-policy-agent/opa v1.7.1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.34.0
	github.com/samber/lo v1.51.0
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	golang.org/x/sync v0.16.0
	golang.org/x/term v0.33.0
	sigs.k8s.io/controller-runtime v0.19.3
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20240806141605-e8a1dd7889d6 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytecodealliance/wasmtime-go/v3 v3.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/containerd/v2 v2.1.4 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v1.0.0-rc.1 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/docker/docker-credential-helpers v0.8.2 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/kyokomi/emoji v2.2.4+incompatible // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/onsi/ginkgo/v2 v2.23.4 // indirect
	github.com/onsi/gomega v1.37.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/peterh/liner v1.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.23.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/procfs v0.17.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tchap/go-patricia/v2 v2.3.3 // indirect
	github.com/vektah/gqlparser/v2 v2.5.30 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yashtewari/glob-intersection v0.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/mod v0.26.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/grpc v1.74.2 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.5.2 // indirect
	oras.land/oras-go/v2 v2.6.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
