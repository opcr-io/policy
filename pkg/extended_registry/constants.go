package extendedregistry

// OCI Annotations
const (
	AnnotationPolicyRegistryType = "org.openpolicyregistry.type"
	PolicyTypePolicy             = "policy"

	AnnotationPolicyRegistryTemplateKind = "org.openpolicyregistry.template.kind"
	TemplateTypePolicy                   = "policy"
	TemplateTypeCICD                     = "cicd"

	AnnotationImageVendor      = "org.opencontainers.image.vendor"
	AnnotationImageAuthors     = "org.opencontainers.image.authors"
	AnnotationImageDescription = "org.opencontainers.image.description"
	AnnotationImageTitle       = "org.opencontainers.image.title"
	AnnotationImageCreated     = "org.opencontainers.image.created"
)
