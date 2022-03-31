package authflow

import (
	"fmt"
	"github.com/99designs/keyring"
	"github.com/brawaru/marct/launcher/accounts"
	xboxAccount "github.com/brawaru/marct/xbox/account"
	"github.com/bwmarrin/snowflake"
	"strconv"
)

type ReadKeyringKeyStep struct {
	Keyring keyring.Keyring // Keyring used to get the key.
	s       snowflake.Node
}

const StepIDReadKeyringKey = "read_keyring_key"

func (r *ReadKeyringKeyStep) ID() string {
	return StepIDReadKeyringKey
}

func (r *ReadKeyringKeyStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	k, err := xboxAccount.RandomKey()
	if err != nil {
		return fmt.Errorf("cannot create secure key: %w", err)
	}
	is := state.IntermediateState
	is.key = k
	is.XboxAccountProperties.KID = strconv.FormatInt(r.s.Generate().Int64(), 16)
	return nil
}

func (r *ReadKeyringKeyStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	i := state.IntermediateState
	kid := prefixedID(i.XboxAccountProperties.KID)
	ki, err := r.Keyring.Get(kid)
	if err != nil {
		return fmt.Errorf("cannot read saved secure key %s: %w", kid, err)
	}
	i.key = ki.Data
	return nil
}
