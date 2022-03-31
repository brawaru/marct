package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/brawaru/marct/launcher/accounts"
	"time"
)

type AuthData struct {
	MsftAccessToken          string    `json:"msftAccessToken"`
	MsftRefreshToken         string    `json:"msftRefreshToken"`
	MsftAccessTokenExpiresAt time.Time `json:"msftAccessTokenExpiresAt"`
	XBLToken                 string    `json:"xblToken"`
	UserHash                 string    `json:"userHash"`
	XSTSToken                string    `json:"xstsToken"`
	MinecraftToken           string    `json:"minecraftToken"`
	MinecraftTokenExpiresAt  time.Time `json:"minecraftAccessTokenExpiresAt"`
}

func createAESGCM(key []byte) (cipher.AEAD, error) {
	block, cipherErr := aes.NewCipher(key)
	if cipherErr != nil {
		return nil, cipherErr
	}

	gcm, gcmErr := cipher.NewGCM(block)
	if gcmErr != nil {
		return nil, gcmErr
	}

	return gcm, nil
}

func ReadAuthData(account accounts.Account, key []byte) (data AuthData, err error) {
	if account.AuthData == nil {
		err = errors.New("account AuthData is nil")
		return
	}

	enc, decodeErr := base64.StdEncoding.DecodeString(*account.AuthData)
	if decodeErr != nil {
		err = fmt.Errorf("cannot decode AuthData: %w", decodeErr)
		return
	}

	gcm, gcmCreateErr := createAESGCM(key)
	if gcmCreateErr != nil {
		err = fmt.Errorf("cannot init AES GCM: %w", gcmCreateErr)
		return
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	bytes, openErr := gcm.Open(nil, nonce, ciphertext, nil)
	if openErr != nil {
		err = fmt.Errorf("cannot decrypt AuthData: %w", openErr)
		return
	}

	if unmarshalErr := json.Unmarshal(bytes, &data); unmarshalErr != nil {
		err = fmt.Errorf("cannot unmarshal AuthData: %w", unmarshalErr)
		return
	}

	return
}

func WriteAuthData(account *accounts.Account, data AuthData, key []byte) error {
	bytes, marshalErr := json.Marshal(data)

	if marshalErr != nil {
		return fmt.Errorf("cannot marshal AuthData: %w", marshalErr)
	}

	gcm, gcmCreateErr := createAESGCM(key)
	if gcmCreateErr != nil {
		return fmt.Errorf("cannot init AES GCM: %w", gcmCreateErr)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, randErr := rand.Read(nonce); randErr != nil {
		return fmt.Errorf("failed to create nonce: %w", randErr)
	}

	ciphertext := gcm.Seal(nonce, nonce, bytes, nil)

	enc := base64.StdEncoding.EncodeToString(ciphertext)
	account.AuthData = &enc

	return nil
}

// RandomKey returns a new byte slice of random bytes to use as a key for Xbox account.
func RandomKey() (key []byte, err error) {
	key = make([]byte, 32)
	_, err = rand.Read(key)

	return
}
