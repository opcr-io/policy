module github.com/aserto-dev/policy

go 1.16

replace github.com/open-policy-agent/opa => github.com/open-policy-agent/opa v0.25.2

replace github.com/aserto-dev/aserto-runtime => ../aserto-runtime

replace github.com/aserto-dev/clui => ../clui

require (
	github.com/alecthomas/kong v0.2.17
	github.com/aserto-dev/aserto-runtime v0.0.21
	github.com/aserto-dev/calc-version v1.1.4 // indirect
	github.com/aserto-dev/clui v0.0.1
	github.com/aserto-dev/go-eds v0.0.22
	github.com/aserto-dev/go-lib v0.2.74
	github.com/aserto-dev/mage-loot v0.2.37
	github.com/containerd/containerd v1.5.2
	github.com/distribution/distribution v2.7.1+incompatible
	github.com/docker/cli v20.10.8+incompatible // indirect
	github.com/docker/docker v20.10.8+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/magefile/mage v1.11.0
	github.com/open-policy-agent/opa v0.28.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.23.0
	github.com/smartystreets/assertions v1.1.1 // indirect
	github.com/spf13/viper v1.8.1
	github.com/urfave/cli/v2 v2.3.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	oras.land/oras-go v0.4.0
	sigs.k8s.io/controller-runtime v0.9.5
)
