package policytemplates

import (
	"io/fs"

	"github.com/aserto-dev/go-grpc/aserto/api/v1"
)

type PolicyTemplates interface {
	ListRepos(org, tag string) (map[string]*api.RegistryRepoTag, error)
	Load(userRef string) (fs.FS, error)
}
