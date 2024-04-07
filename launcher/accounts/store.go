package accounts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/validfile"
	"github.com/rogpeppe/go-internal/lockedfile"
)

type Account struct {
	Type          string            `json:"type"`               // Account type (e.g. "Xbox").
	ID            string            `json:"id"`                 // Account ID.
	AuthData      *string           `json:"authData,omitempty"` // Encoded account data.
	Properties    map[string]string `json:"properties"`         // Account properties (like avatar or display name).
	Authorization *Authorization    `json:"-"`                  // Authorization data.
}

type Store struct {
	Accounts        map[string]*Account `json:"accounts"`                  // Accounts mapped by their IDs.
	SelectedAccount string              `json:"selectedAccount,omitempty"` // ID of the selected account.
}

func (s *Store) AddAccount(account Account) {
	s.Accounts[account.ID] = &account
}

func (s *Store) RemoveAccount(account Account) {
	delete(s.Accounts, account.ID)
	if s.SelectedAccount == account.ID {
		s.SelectedAccount = ""
	}
}

// GetSelectedAccount returns the reference to the selected account or nil if no account is selected.
func (s *Store) GetSelectedAccount() *Account {
	if s.SelectedAccount == "" {
		return nil
	}

	for _, account := range s.Accounts {
		if account.ID == s.SelectedAccount {
			return account
		}
	}

	return nil
}

type StoreFile struct {
	Store
	file *lockedfile.File // File open.
}

// Save saves the file.
func (f *StoreFile) Save() error {
	if f.file == nil {
		return errors.New("file is closed")
	}

	buf := new(bytes.Buffer)

	if marshalErr := json.NewEncoder(buf).Encode(f.Store); marshalErr != nil {
		return fmt.Errorf("marshal error: %w", marshalErr)
	}

	if truncErr := f.file.Truncate(0); truncErr != nil {
		return fmt.Errorf("cannot truncate file: %w", truncErr)
	}

	if _, seekErr := f.file.Seek(0, 0); seekErr != nil {
		return fmt.Errorf("cannot seek in file: %w", seekErr)
	}

	if _, writeErr := io.Copy(f.file, buf); writeErr != nil {
		return fmt.Errorf("write failure: %w", writeErr)
	}

	//if _, writeErr := f.file.Write(bytes); writeErr != nil {
	//	return fmt.Errorf("write failure: %w", writeErr)
	//}

	if syncErr := f.file.Sync(); syncErr != nil {
		return fmt.Errorf("sync error: %w", syncErr)
	}

	return nil
}

func (f *StoreFile) Close() error {
	if f.file == nil {
		return errors.New("file is closed")
	}

	if saveErr := f.Save(); saveErr != nil {
		return saveErr
	}

	if fileCloseErr := f.file.Close(); fileCloseErr != nil {
		return fmt.Errorf("failed to close store file: %w", fileCloseErr)
	}

	f.file = nil

	return nil
}

func OpenStoreFile(filepath string) (*StoreFile, error) {
	exists, checkErr := validfile.FileExists(filepath)
	if checkErr != nil {
		return nil, fmt.Errorf("failed to check store file %s existence: %w", filepath, checkErr)
	}

	file, openErr := lockedfile.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0660)
	if openErr != nil {
		return nil, fmt.Errorf("failed to open store file %s: %w", filepath, openErr)
	}

	var s Store

	if exists {
		if readErr := json.NewDecoder(file).Decode(&s); readErr != nil {
			defer utils.DClose(file)
			return nil, fmt.Errorf("failed to read store file %s: %w", filepath, readErr)
		}
	} else {
		// initialise empty store if file did not exist before
		s = Store{
			Accounts:        map[string]*Account{},
			SelectedAccount: "",
		}
	}

	r := &StoreFile{
		file:  file,
		Store: s,
	}

	if !exists {
		if saveErr := r.Save(); saveErr != nil {
			return nil, saveErr
		}
	}

	return r, nil
}
