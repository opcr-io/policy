module github.com/opcr-io/policy

go 1.16

// replace github.com/aserto-dev/clui => ../clui

// replace github.com/aserto-dev/mage-loot => ../mage-loot

// replace github.com/aserto-dev/aserto-runtime => ../aserto-runtime

require (
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/Microsoft/hcsshim v0.8.20 // indirect
	github.com/alecthomas/kong v0.2.17
	github.com/aserto-dev/aserto-runtime v0.0.24-0.20210902103415-12e69833e705
	github.com/aserto-dev/clui v0.1.4
	github.com/aserto-dev/go-utils v0.0.8
	github.com/aserto-dev/mage-loot v0.4.6
	github.com/bugsnag/bugsnag-go v1.0.5-0.20150529004307-13fd6b8acda0 // indirect
	github.com/containerd/containerd v1.5.3
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/golang/snappy v0.0.4-0.20210608040537-544b4180ac70 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/go-containerregistry v0.6.0
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/klauspost/compress v1.13.4 // indirect
	github.com/magefile/mage v1.11.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/open-policy-agent/opa v0.31.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.2-0.20210819154149-5ad6f50d6283
	github.com/pkg/errors v0.9.1
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/rs/zerolog v1.23.0
	github.com/smartystreets/assertions v1.1.1 // indirect
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.8.1
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210820121016-41cdb8703e55 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d
	google.golang.org/genproto v0.0.0-20210811021853-ddbe55d93216 // indirect
	gopkg.in/yaml.v2 v2.4.0
	oras.land/oras-go v0.4.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.9.5
)
