package authflow

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/brawaru/marct/launcher/accounts"
)

const StepIDRequestUsername = "request_username"

type RequestUsernameStep struct {
	UsernameRequestHandler UsernameRequestHandler
}

func (s *RequestUsernameStep) ID() string {
	return StepIDRequestUsername
}

func randomID() (string, error) {
	hash := sha1.New()

	if _, err := io.CopyN(hash, rand.Reader, 256); err != nil {
		return "", fmt.Errorf("cannot generate random bytes: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (s *RequestUsernameStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	if id, err := randomID(); err != nil {
		return fmt.Errorf("generate random ID: %w", err)
	} else {
		state.IntermediateState.MinecraftAccountProperties.ID = id
		state.Account.ID = id
	}

	username, err := s.UsernameRequestHandler()
	if err != nil {
		return fmt.Errorf("request username: %w", err)
	}
	state.IntermediateState.MinecraftAccountProperties.Username = username

	return nil
}

func (s *RequestUsernameStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	if state.IntermediateState.MinecraftAccountProperties.Username == "" {
		return s.Authorize(state)
	}

	return nil
}
