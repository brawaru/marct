package authflow

import (
	"fmt"
	"github.com/brawaru/marct/launcher/accounts"
	minecraftAPI "github.com/brawaru/marct/minecraft/api"
	"github.com/brawaru/marct/utils/slices"
)

type EntitlementsCheckStep struct {
	RequiredEntitlements []string
}

const StepIDEntitlementsCheck = "entitlements_check"

func (s *EntitlementsCheckStep) ID() string {
	return StepIDEntitlementsCheck
}

type EntitlementMissingError struct {
	Entitlement string
}

func (e *EntitlementMissingError) Error() string {
	return fmt.Sprintf("user is missing entitlement %s", e.Entitlement)
}

func (e *EntitlementMissingError) Is(target error) bool {
	t, ok := target.(*EntitlementMissingError)
	return ok && ((t != nil) == (e != nil)) &&
		(t.Entitlement == "" || t.Entitlement == e.Entitlement)
}

func (s *EntitlementsCheckStep) check(state *accounts.FlowState[IntermediateState]) error {
	is := state.IntermediateState

	mcAPI := minecraftAPI.NewAuthorizedAPI(is.MinecraftToken)

	entitlements, entitlementsFetchErr := mcAPI.GetEntitlements()
	if entitlementsFetchErr != nil {
		return fmt.Errorf("cannot fetch entitlements: %w", entitlementsFetchErr)
	}

	// FIXME: ReadSignature() is broken :(
	//_, signatureReadErr := entitlements.ReadSignature()
	//if signatureReadErr != nil {
	//	return signatureReadErr
	//}

	for _, entitlementName := range s.RequiredEntitlements {
		if !slices.Some(entitlements.Items, func(item *minecraftAPI.Entitlement, _ int, _ []minecraftAPI.Entitlement) bool {
			return item.Name == entitlementName
		}) {
			return &EntitlementMissingError{
				Entitlement: entitlementName,
			}
		}
	}

	return nil
}

func (s *EntitlementsCheckStep) Authorize(state *accounts.FlowState[IntermediateState]) error {
	return s.check(state)
}

func (s *EntitlementsCheckStep) Refresh(state *accounts.FlowState[IntermediateState]) error {
	return s.check(state)
}
