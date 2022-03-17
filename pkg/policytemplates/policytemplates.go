package policytemplates

import "io/fs"

type PolicyTemplates interface {
	ListRepos() ([]string, error)
	Load(string) (fs.FS, error)
}
