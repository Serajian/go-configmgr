# Changelog

All notable changes to this project will be documented in this file.  
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),  
and this project adheres to [Semantic Versioning](https://semver.org/).

---

## [v1.0.0] - 2025-08-30
### Added
- Load configuration from multiple sources:
    - JSON / YAML files
    - `.env` files
    - System environment variables
- Profile-based overrides (e.g., `config-dev.yaml`, `.env.prod`)
- Default values via struct tags (`default:"value"`)
- Validation via [go-playground/validator](https://github.com/go-playground/validator)
- Key normalization (all keys uppercase)
- Export configuration to JSON / YAML
- Encrypted configs (AES-GCM: `.yaml.enc`, `.json.enc`)
- Test utilities (`NewTestConfig`)
- Pluggable logging
- CLI tool (`configctl`) to inspect configs

---
