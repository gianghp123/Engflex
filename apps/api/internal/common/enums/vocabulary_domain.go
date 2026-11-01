package enums

// VocabularyDomain groups technical vocabulary (vocabulary_items.domain).
// Open set: when a domain is added, extend these constants and the
// chk_vocabulary_items_domain CHECK together.
type VocabularyDomain string

const (
	VocabularyDomainBackendDB   VocabularyDomain = "backend_db"
	VocabularyDomainDistributed VocabularyDomain = "distributed_systems"
	VocabularyDomainDevOpsCloud VocabularyDomain = "devops_cloud"
	VocabularyDomainFrontendUI  VocabularyDomain = "frontend_ui"
	VocabularyDomainAIML        VocabularyDomain = "ai_ml"
)
