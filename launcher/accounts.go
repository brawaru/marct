package launcher

import (
	"fmt"
	"path/filepath"

	"github.com/99designs/keyring"
	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/utils/slices"
)

func (w *Instance) OpenAccountsStore() (*accounts.StoreFile, error) {
	return accounts.OpenStoreFile(filepath.Join(w.Path, filepath.FromSlash(accountsPath)))
}

type SelectBackendFunc func(backends []keyring.BackendType) (keyring.BackendType, error)

type KeyringOpenOptions struct {
	Backend    keyring.BackendType // Keyring backend to use.
	PromptFunc keyring.PromptFunc  // PromptFunc is a function that is called when the keyring needs a password to unlock.
	PassCmd    string
	PassDir    string
}

func SelectKeyringBackend(selectFunc SelectBackendFunc) (keyring.BackendType, error) {
	backends := keyring.AvailableBackends()
	bt, err := selectFunc(backends)

	if err != nil {
		return keyring.InvalidBackend, fmt.Errorf("select keyring backend: %w", err)
	}

	if !slices.Includes(backends, bt) {
		return keyring.InvalidBackend, fmt.Errorf("invalid keyring backend: %s", bt)
	}

	return bt, nil
}

func (w *Instance) OpenKeyring(opts KeyringOpenOptions) (keyring.Keyring, error) {
	if !slices.Includes(keyring.AvailableBackends(), opts.Backend) {
		return nil, fmt.Errorf("backend %s is not available", opts.Backend)
	}

	return keyring.Open(keyring.Config{
		KeychainPasswordFunc: opts.PromptFunc,
		FilePasswordFunc:     opts.PromptFunc,
		ServiceName:          "marct",
		WinCredPrefix:        "marct",
		KeychainName:         "marct",
		PassPrefix:           "marct/",
		AllowedBackends:      []keyring.BackendType{opts.Backend},
		FileDir:              w.Path,
		PassCmd:              opts.PassCmd,
		PassDir:              opts.PassDir,
	})
}
