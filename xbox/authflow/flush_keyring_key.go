package authflow

import (
	"fmt"
	"github.com/99designs/keyring"
	"github.com/brawaru/marct/launcher/accounts"
)

type FlushKeyringKeyStep struct {
	Keyring keyring.Keyring
}

const StepIDFlushKeyringKey = "flush_keyring_key"

func (s *FlushKeyringKeyStep) ID() string {
	return StepIDFlushKeyringKey
}

func (s *FlushKeyringKeyStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	i := state.IntermediateState

	accountID := i.XboxAccountProperties.KID
	if err := s.Keyring.Set(keyring.Item{
		Key:  prefixedID(accountID),
		Data: i.key,
	}); err != nil {
		return fmt.Errorf("cannot save key for %s: %w", accountID, err)
	}

	return nil
}

func (s *FlushKeyringKeyStep) Refresh(*accounts.FlowState[IntermediateState]) error {
	return nil // we expect the key to stay the same
}
