module github.com/opcr-io/policy/oci

go 1.19

replace oras.land/oras-go/v2 => github.com/opcr-io/oras-go/v2 v2.0.0-20230921121537-80bf1a01f1b6

require (
	github.com/containerd/containerd v1.6.19
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0-rc4
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.29.1
	oras.land/oras-go/v2 v2.0.0
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.29.1 // indirect
)
