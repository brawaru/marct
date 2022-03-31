package accounts

import (
	"errors"
	"fmt"
)

// This is kinda inspired by MultiMC

// IntermediateState represents an intermediate state within authentication flow, it can be anything that helps get
// authentication done, and can be modified throughout the flow.
//
// For example, Xbox authentication stores Authentication data in it, which gets modified with every step adding new
// data, such as Xbox Live token, Minecraft token. In the end this authentication data gets stored with new account.
type IntermediateState interface{}

// FlowState represents a mutable structure that represents current authentication flow state. It is always passed by
// reference and supposed to be modified by every authentication step.
type FlowState[I IntermediateState] struct {
	Account           *Account // Account which being authenticated.
	IntermediateState *I       // Shared state between every step of the flow.
}

// FlowStep represents an authentication step.
type FlowStep[I IntermediateState] interface {
	ID() string                          // ID returns flow step identifier.
	Authorize(state *FlowState[I]) error // Authorize performs authorization.
	Refresh(state *FlowState[I]) error   // Refresh refreshes the existing token.
}

// AuthFlow represents an authentication flow. It is used to authenticate new or refresh existing accounts.
type AuthFlow[I IntermediateState] struct {
	AccountType  string        // Accounts type this flow creates.
	Steps        []FlowStep[I] // Every step of the flow to run sequentially.
	InitialState I             // Initial state for every flow run.
}

// AddStep adds a new step to the flow.
func (f *AuthFlow[I]) AddStep(step FlowStep[I]) {
	f.Steps = append(f.Steps, step)
}

// StepError represents an error that occurred during the authentication flow.
type StepError struct {
	StepID string // Identifier of the step that failed.
	Err    error  // Error that occurred during running of this step.
}

func (s *StepError) Error() string {
	return fmt.Sprintf("failed to run step %s: %s", s.StepID, s.Err.Error())
}

func (s *StepError) Unwrap() error {
	return s.Err
}

func (s *StepError) Is(target error) bool {
	t, ok := target.(*StepError)
	return ok && (t.StepID == "" || t.StepID == s.StepID) &&
		(t.Err == nil || errors.Is(s.Err, t.Err))
}

// CreateAccount creates a new account of type specified by the flow.
func (f *AuthFlow[I]) CreateAccount() (Account, error) {
	account := Account{Type: f.AccountType}

	initialState := f.InitialState

	state := FlowState[I]{
		Account:           &account,
		IntermediateState: &initialState,
	}

	for _, step := range f.Steps {
		if stepErr := step.Authorize(&state); stepErr != nil {
			return account, &StepError{
				StepID: step.ID(),
				Err:    stepErr,
			}
		}
	}

	return account, nil
}

// RefreshAccount refreshes an existing account of type specified by the flow.
func (f *AuthFlow[I]) RefreshAccount(account *Account) error {
	initialState := f.InitialState

	state := FlowState[I]{
		Account:           account,
		IntermediateState: &initialState,
	}

	for _, step := range f.Steps {
		if stepErr := step.Refresh(&state); stepErr != nil {
			return &StepError{
				StepID: step.ID(),
				Err:    stepErr,
			}
		}
	}

	return nil
}
