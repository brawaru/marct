package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/brawaru/marct/launcher/accounts"
)

// AuthData stores the authentication data for Xbox Minecraft account.
type AuthData struct {
	MsftAccessToken          string    `json:"msftAccessToken"`               // Microsoft access token used to access Xbox services.
	MsftRefreshToken         string    `json:"msftRefreshToken"`              // Microsoft refresh token used to refresh access token after it has expired.
	MsftAccessTokenExpiresAt time.Time `json:"msftAccessTokenExpiresAt"`      // Time after which Microsoft token should be considered expired.
	XBLToken                 string    `json:"xblToken"`                      // Xbox Live token.
	UserHash                 string    `json:"userHash"`                      // Xbox Live user hash.
	XSTSToken                string    `json:"xstsToken"`                     // Xbox Live Token Service token.
	MinecraftToken           string    `json:"minecraftToken"`                // Minecraft token used to call Minecraft APIs and to log in into the game.
	MinecraftTokenExpiresAt  time.Time `json:"minecraftAccessTokenExpiresAt"` // Time after which Minecraft token should be considered expired.
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
