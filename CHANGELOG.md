# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2024-01-XX

### Added
- Initial release of the Tama Terraform provider
- **Neural Resources:**
  - `tama_space` resource for managing neural spaces
  - `tama_space` data source for referencing existing spaces
- **Sensory Resources:**
  - `tama_source` resource for managing AI service sources
  - `tama_source` data source for referencing existing sources
  - `tama_model` resource for managing AI models
  - `tama_model` data source for referencing existing models
  - `tama_limit` resource for managing rate limits
  - `tama_limit` data source for referencing existing limits
- Provider configuration with authentication support
- Support for environment variable configuration (`TAMA_BASE_URL`, `TAMA_API_KEY`)
- Comprehensive examples for all resources and data sources
- Import support for all resources
- Integration with `github.com/upmaru/tama-go` client library

### Features
- **Spaces**: Create and manage neural spaces with types (root, component)
- **Sources**: Configure AI service providers with endpoints and API keys
- **Models**: Define available AI models from sources with identifiers and paths
- **Limits**: Set rate limiting rules with flexible time units and counts
- **Authentication**: Secure API key-based authentication
- **Import**: Import existing resources using their IDs

### Documentation
- Complete provider documentation with examples
- Resource and data source reference documentation
- Quick start guide and usage examples
- Multi-provider setup examples (OpenAI, Mistral, Anthropic)
- Relationship documentation between resources

### Technical Details
- Built on Terraform Plugin Framework v1.15.0
- Go 1.23+ support
- Comprehensive error handling and validation
- Structured logging support
- Plan modifiers for computed and sensitive attributes