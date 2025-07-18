# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Model Parameters Support**: Added `parameters` attribute to `tama_model` resource and data source
  - Supports flexible model configuration through JSON parameters
  - Compatible with tama-go client library v0.1.12+
  - Enables configuration of temperature, reasoning effort, max tokens, and other model-specific settings
  - Supports complex nested JSON structures including objects, arrays, and all JSON data types
  - Backward compatible - existing configurations continue to work unchanged
  - Optional field with comprehensive error handling and validation

### Enhanced
- **Model Resource**: Enhanced with parameter support for advanced model configuration
- **Model Data Source**: Enhanced to return model parameters from API responses
- **Import Functionality**: Model imports now include parameter configuration when available
- **Documentation**: Added comprehensive parameter usage guide and examples

### Technical
- JSON parameter serialization/deserialization with proper error handling
- Support for any valid JSON structure via `map[string]any`
- Graceful handling of null/empty parameters
- Enhanced test coverage with 13+ new test cases covering all parameter scenarios

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