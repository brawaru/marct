package xbox

import "fmt"

const ( // well known xErrs
	xErrAccountDoesntExist    = 2148916233 // Xbox account is not created.
	xErrUnavailableCountry    = 2148916235 // Xbox is not available in user's country.
	xErrAdultConsentRequired  = 2148916236 // Adult consent required (South Korea).
	xErrAdultConsentRequired2 = 2148916237 // Adult consent required (South Korea).
	xErrManagedAccount        = 2148916238 // The account is managed but not added to the family (underage).
)

var (
	XErrAccountDoesntExist    = &XTokenError{XErr: xErrAccountDoesntExist}
	XErrUnavailableCountry    = &XTokenError{XErr: xErrUnavailableCountry}
	XErrAdultConsentRequired  = &XTokenError{XErr: xErrAdultConsentRequired}
	XErrAdultConsentRequired2 = &XTokenError{XErr: xErrAdultConsentRequired2}
	XErrManagedAccount        = &XTokenError{XErr: xErrManagedAccount}
)

type XTokenError struct {
	Identity string `json:"Identity"`
	XErr     int64  `json:"XErr"`
	Message  string `json:"Message"`
	Redirect string `json:"Redirect"`
}

func (e *XTokenError) Error() string {
	var message string
	if len(e.Message) != 0 {
		message = e.Message
	} else {
		switch e.XErr {
		case xErrAccountDoesntExist:
			message = "Xbox account is not created."
		case xErrUnavailableCountry:
			message = "Xbox is not available in the country."
		case xErrAdultConsentRequired:
			fallthrough
		case xErrAdultConsentRequired2:
			message = "adult consent is required for child account."
		case xErrManagedAccount:
			message = "the account is managed and needs to be added to the family."
		default:
			message = "unknown error."
		}
	}

	return fmt.Sprintf("XErr %v: %s", e.XErr, message)
}

func (e *XTokenError) Is(target error) bool {
	t, ok := target.(*XTokenError)
	return ok && (t.Redirect == "" || e.Redirect == t.Redirect) &&
		(t.Identity == "" || e.Identity == t.Identity) &&
		(t.XErr == 0 || e.XErr == t.XErr) &&
		(t.Message == "" || e.Message == t.Message)
}
