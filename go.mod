module github.com/aserto-dev/policy

go 1.16

replace github.com/open-policy-agent/opa => github.com/open-policy-agent/opa v0.25.2

replace github.com/aserto-dev/aserto-runtime => ../aserto-runtime

replace github.com/aserto-dev/clui => ../clui

require (
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/alecthomas/kong v0.2.17
	github.com/aserto-dev/aserto-runtime v0.0.21
	github.com/aserto-dev/clui v0.0.1
	github.com/aserto-dev/go-eds v0.0.22
	github.com/aserto-dev/go-lib v0.2.74
	github.com/aserto-dev/mage-loot v0.2.37
	github.com/containerd/containerd v1.5.2
	github.com/dustin/go-humanize v1.0.0
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/klauspost/compress v1.13.0 // indirect
	github.com/magefile/mage v1.11.0
	github.com/open-policy-agent/opa v0.28.0
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.23.0
	github.com/smartystreets/assertions v1.1.1 // indirect
	github.com/spf13/cobra v1.2.1 // indirect
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	gopkg.in/yaml.v2 v2.4.0
	oras.land/oras-go v0.4.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.9.5
)
