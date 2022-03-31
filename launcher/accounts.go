package launcher

import (
	"fmt"
	"github.com/99designs/keyring"
	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/utils/slices"
	"path/filepath"
)

func (w *Instance) OpenAccountsStore() (*accounts.StoreFile, error) {
	return accounts.OpenStoreFile(filepath.Join(w.Path, filepath.FromSlash(accountsPath)))
}

func (w *Instance) OpenKeyring(promptFunc keyring.PromptFunc) (keyring.Keyring, error) {
	// the reason we put it on working directory is that we want to make it configurable in the future

	settings := w.Settings.Keyring
	bannedBackends := settings.BanBackends
	var allowedBackends []keyring.BackendType
	for _, backendType := range keyring.AvailableBackends() {
		if slices.Includes(bannedBackends, backendType) {
			fmt.Printf("keyring-open: skipping banned backend %s\n", backendType)
		} else {
			allowedBackends = append(allowedBackends, backendType)
		}
	}

	return keyring.Open(keyring.Config{
		KeychainPasswordFunc: promptFunc,
		FilePasswordFunc:     promptFunc,
		ServiceName:          "marct",
		WinCredPrefix:        "marct",
		KeychainName:         "marct",
		PassPrefix:           "marct/",
		AllowedBackends:      allowedBackends,
		FileDir:              w.Path,
		PassCmd:              settings.PassCmd,
		PassDir:              settings.PassDir,
	})
}
