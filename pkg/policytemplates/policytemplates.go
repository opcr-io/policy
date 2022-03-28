package policytemplates

import "io/fs"

type PolicyTemplates interface {
	ListRepos(org, tag string) ([]string, error)
	Load(userRef string) (fs.FS, error)
}
