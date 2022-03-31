package launcher

import (
	"encoding/json"
	"fmt"
)

type RuleAction string

const (
	Allow    RuleAction = "allow"
	Disallow RuleAction = "disallow"
)

func (r *RuleAction) IsValid() error {
	switch *r {
	case Allow:
		return nil
	case Disallow:
		return nil
	default:
		return fmt.Errorf("%#v is not valid value for RuleAction", r)
	}
}

func (r *RuleAction) String() (string, error) {
	if err := r.IsValid(); err != nil {
		return "", err
	}

	return string(*r), nil
}

func (r *RuleAction) UnmarshalJSON(i []byte) error {
	var v string

	err := json.Unmarshal(i, &v)
	if err != nil {
		return err
	}

	res := RuleAction(v)

	if err := res.IsValid(); err != nil {
		return err
	}

	*r = res

	return nil
}

func (r *RuleAction) MarshalJSON() ([]byte, error) {
	if err := r.IsValid(); err != nil {
		return nil, err
	}

	return json.Marshal(string(*r))
}

func (r *RuleAction) Invert() RuleAction {
	switch *r {
	case Allow:
		return Disallow
	case Disallow:
		return Allow
	default:
		return ""
	}
}

type Feature string

const (
	FeatDemoUser         Feature = "is_demo_user"
	FeatCustomResolution Feature = "has_custom_resolution"
)

type Rule struct {
	Action   RuleAction        `json:"action"`
	OS       *OS               `json:"os,omitempty"`
	Features *map[Feature]bool `json:"features,omitempty"`
}

// Decide decides the action to take based on basic rules.
// It does not take features into account, for this DecideExtensively must be used.
func (r *Rule) Decide() RuleAction {
	if r.OS != nil && !r.OS.Matches() {
		return r.Action.Invert()
	}

	return r.Action
}

// DecideExtensively decides the action to take based on the rules.
// It does check basic rules, so it is pointless to call Decide together with it.
func (r *Rule) DecideExtensively(featSet map[Feature]bool) RuleAction {
	if basicDecision := r.Decide(); basicDecision != r.Action {
		return basicDecision
	}

	if r.Features != nil {
		if featSet == nil {
			return r.Action.Invert()
		}

		for feature, expectedState := range *r.Features {
			featState, isSet := featSet[feature]

			if expectedState != (isSet && featState) {
				return r.Action.Invert()
			}
		}
	}

	return r.Action
}

type Rules []Rule

func (r *Rules) MatchesExtensively(featSet map[Feature]bool) bool {
	if r == nil {
		return true
	}

	for _, rule := range *r {
		if rule.DecideExtensively(featSet) == Disallow {
			return false
		}
	}

	return true
}

func (r *Rules) Matches() bool {
	if r == nil {
		return true
	}

	anyMatched := false

	for _, rule := range *r {
		if rule.Decide() == Disallow {
			return false
		} else {
			anyMatched = true
		}
	}

	return anyMatched
}
