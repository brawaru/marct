package authflow

import "github.com/brawaru/marct/xbox"

func findUserHash(claims xbox.XDisplayClaims) (userHash string, found bool) {
	xuiClaim, xuiClaimExists := claims["xui"]
	if !xuiClaimExists {
		return
	}

	for _, e := range xuiClaim {
		userHashClaim, it := e.(map[string]interface{})

		if !it {
			continue
		}

		u, uhsFound := userHashClaim["uhs"].(string)

		if !uhsFound {
			continue
		}

		if it {
			userHash = u
			found = true
			break
		}
	}

	return
}

func prefixedID(id string) string {
	return "xbox:" + id
}
