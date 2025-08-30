# Release Notes

## v1.0.0 - Initial Release 🎉

✨ First stable release of **go-configmgr**

### Features
- ✅ Load configuration from multiple sources:
    - JSON / YAML files
    - `.env` files
    - System environment variables
- ✅ Profile-based overrides (e.g., `config-dev.yaml`, `.env.prod`)
- ✅ Default values via struct tags (`default:"value"`)
- ✅ Validation via [go-playground/validator](https://github.com/go-playground/validator)
- ✅ Key normalization (all keys uppercase)
- ✅ Export configuration to JSON / YAML
- ✅ Encrypted configs (AES-GCM: `.yaml.enc`, `.json.enc`)
- ✅ Test utilities (`NewTestConfig`)
- ✅ Pluggable logging
- ✅ CLI tool (`configctl`) to inspect configs

### Examples
- Simple `.env` loader
- YAML + struct with defaults
- Profile-based config (`APP_ENV=dev`)
- Encrypted config file

### Roadmap
- [ ] Hot reload (fsnotify)
- [ ] Integration with secret managers (Vault, AWS, GCP)
- [ ] More real-world microservice examples
