module github.com/opcr-io/policy/oci

go 1.19

replace oras.land/oras-go/v2 v2.2.0 => github.com/opcr-io/oras-go/v2 v2.0.0-20230331095552-41df01dce636

require (
	github.com/containerd/containerd v1.7.2
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0-rc2.0.20221005185240-3a7f492d3f1b
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.29.1
	oras.land/oras-go/v2 v2.2.0
)

require (
	github.com/AdaLogics/go-fuzz-headers v0.0.0-20230106234847-43070de90fa1 // indirect
	github.com/cyphar/filepath-securejoin v0.2.3 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	go.opentelemetry.io/otel v1.14.0 // indirect
	go.opentelemetry.io/otel/trace v1.14.0 // indirect
	golang.org/x/sync v0.2.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	google.golang.org/grpc v1.54.0 // indirect
	google.golang.org/protobuf v1.29.1 // indirect
)
