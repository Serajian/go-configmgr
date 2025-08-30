package configmgr

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadEncryptedFile loads and decrypts an encrypted config file (AES-256).
func (cm *ConfigManager) LoadEncryptedFile(path, secret string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	encBytes, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}

	key := sha256.Sum256([]byte(secret))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	if len(encBytes) < gcm.NonceSize() {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encBytes[:gcm.NonceSize()], encBytes[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	ext := getEncryptedExt(path)
	tmp := make(map[string]interface{})

	if ext == ".json.enc" {
		if err = json.Unmarshal(plaintext, &tmp); err != nil {
			return err
		}
	} else if ext == ".yaml.enc" || ext == ".yml.enc" {
		if err = yaml.Unmarshal(plaintext, &tmp); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unsupported encrypted file type: %s", ext)
	}

	for k, v := range tmp {
		cm.data[normalizeKey(k)] = normalizeValue(v)
	}

	return nil
}

func getEncryptedExt(path string) string {
	if strings.HasSuffix(path, ".json.enc") {
		return ".json.enc"
	}
	if strings.HasSuffix(path, ".yaml.enc") {
		return ".yaml.enc"
	}
	if strings.HasSuffix(path, ".yml.enc") {
		return ".yml.enc"
	}
	return filepath.Ext(path) // fallback
}
