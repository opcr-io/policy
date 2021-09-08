module github.com/aserto-dev/policy

go 1.16

replace github.com/aserto-dev/clui => ../clui

// replace github.com/aserto-dev/mage-loot => ../mage-loot

replace github.com/opencontainers/artifacts => github.com/notaryproject/artifacts v0.0.0-20210414030140-c7c701eff45d

replace github.com/notaryproject/nv2 => github.com/notaryproject/nv2 v0.0.0-20210401122849-20e35b6ce1a8

replace github.com/notaryproject/notary/v2 => ../notation-go-lib // github.com/notaryproject/notary/v2 v2.0.0-20210414032403-d1367cc13db7

require (
	github.com/alecthomas/kong v0.2.17
	github.com/aserto-dev/clui v0.1.2
	github.com/aserto-dev/go-utils v0.0.2
	github.com/aserto-dev/mage-loot v0.4.6
	github.com/bytecodealliance/wasmtime-go v0.29.0 // indirect
	github.com/containerd/containerd v1.5.3
	github.com/containers/image/v5 v5.16.0
	github.com/dustin/go-humanize v1.0.0
	github.com/fatih/color v1.12.0 // indirect
	github.com/golang/snappy v0.0.4-0.20210608040537-544b4180ac70 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/magefile/mage v1.11.0
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/open-policy-agent/opa v0.31.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.2-0.20210819154149-5ad6f50d6283
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/rs/zerolog v1.23.0
	github.com/sirupsen/logrus v1.8.1
	github.com/smartystreets/assertions v1.1.1 // indirect
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/theupdateframework/notary v0.7.0
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/urfave/cli v1.22.4
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v2 v2.4.0
	oras.land/oras-go v0.4.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.9.5
)
