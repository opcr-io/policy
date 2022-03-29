package mocks

//go:generate mockgen -destination=mock_registry_client.go -package=mocks github.com/aserto-dev/go-grpc/aserto/registry/v1 RegistryClient
//go:generate mockgen -destination=mock_scc_source.go -package=mocks github.com/aserto-dev/scc-lib/sources Source
