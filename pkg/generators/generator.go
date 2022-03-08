package generators

type GeneratedFilesContent map[string]string

type Generator interface {
	Generate(overwrite bool) error
	GenerateFilesContent() (GeneratedFilesContent, error)
}
