package cmd

import (
	"errors"
	"github.com/99designs/keyring"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/xbox"
	xboxAccount "github.com/brawaru/marct/xbox/account"
	xboxAuthFlow "github.com/brawaru/marct/xbox/authflow"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var accountAddCommand = createCommand(&cli.Command{
	Name: "add",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.accounts-add.usage",
		Other: "Add a new account",
	}),
	Description: locales.Translate(&i18n.Message{
		ID: "command.accounts-add.description",
		Other: "Allows to set up a new account used to log in to the game.\n\n" +
			"This command is a collection of subcommands for every account type.",
	}),
	Subcommands: []*cli.Command{accountAddMicrosoftCommand},
})

func init() {
	accountCommand.Subcommands = append(accountCommand.Subcommands, accountAddCommand)
}

var accountAddMicrosoftCommand = createCommand(&cli.Command{
	Name:    "microsoft",
	Aliases: []string{"msft", "xbox"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.accounts-add-microsoft.usage",
		Other: "A new Microsoft account",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.accounts-add-microsoft.description",
		Other: "Allows to set up a new Microsoft account.",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value("workDir").(*launcher.Instance)

		k, keyringOpenErr := workDir.OpenKeyring(keyringOpenPrompt)
		if keyringOpenErr != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": keyringOpenErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.keyring-open-error",
					Other: "Cannot open your keyring: {{ .Error }}",
				},
			}), 1)
		}

		accountKey, keyCreateErr := xboxAccount.RandomKey()

		if keyCreateErr != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": keyCreateErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.failed-to-generate-key",
					Other: "Cannot generate a key to secure your account data: {{ .Error }}",
				},
			}), 1)
		}

		authFlow := xboxAuthFlow.CreateAuthFlow(&xboxAuthFlow.Options{
			Keyring:           k,
			DeviceAuthHandler: xboxDeviceAuthPrompt,
		})

		//if flowCreateErr != nil {
		//	return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
		//		TemplateData: map[string]string{
		//			"Error": flowCreateErr.Error(),
		//		},
		//		DefaultMessage: &i18n.Message{
		//			ID:    "command.accounts-add-microsoft.error.failed-to-create-flow",
		//			Other: "Cannot create authentication flow: {{ .Error }}",
		//		},
		//	}), 1)
		//}

		accountsStore, openErr := workDir.OpenAccountsStore()

		if openErr != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": openErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.store-open-error",
					Other: "Failed to open the account store: {{ .Error }}",
				},
			}), 1)
		}

		defer utils.DClose(accountsStore)

		account, accountCreateErr := authFlow.CreateAccount()

		if accountCreateErr != nil {
			// FIXME: requires better error handling
			return cli.Exit(translateAccountCreationError(accountCreateErr), 1)
		}

		if keyringSaveErr := k.Set(keyring.Item{
			Key:  "xbox:" + account.ID,
			Data: accountKey,
		}); keyringSaveErr != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": keyringSaveErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.failed-to-save-account-key",
					Other: "Cannot save your account key in keyring: {{ .Error }}",
				},
			}), 1)
		}

		accountsStore.Store.Accounts[account.ID] = account

		if storeSaveErr := accountsStore.Save(); storeSaveErr != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": storeSaveErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.account-store-save-failed",
					Other: "Could not authorize you: failed to save account store; {{ .Error }}",
				},
			}), 1)
		}

		println(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"Username": account.Authorization.UserName,
			},
			DefaultMessage: &i18n.Message{
				ID:    "command.accounts-add-microsoft.welcome",
				Other: "Logged in as {{ .Username }} (⌐■_■)",
			},
		}))

		return nil
	},
})

func translateAccountCreationError(err error) string {
	var stepErr *accounts.StepError

	if errors.As(err, &stepErr) {
		switch stepErr.StepID {
		case xboxAuthFlow.StepIDReadData:
			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": stepErr.Err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.read-data-failed",
					Other: "Cannot initialise a new account: {{ .Error }}",
				},
			})

		case xboxAuthFlow.StepIDDeviceAuth:
			var tokenErr *xbox.TokenError

			if errors.As(stepErr.Err, &tokenErr) {
				if tokenErr.Is(xbox.ExpiredTokenErr) {
					return locales.Translate(&i18n.Message{
						ID:    "command.accounts-add-microsoft.error.token-expired",
						Other: "Authorization request has expired, please try again",
					})
				}

				if tokenErr.Is(xbox.AuthorizationDeclinedErr) {
					return locales.Translate(&i18n.Message{
						ID:    "command.accounts-add-microsoft.error.authorization-declined",
						Other: "You have declined the authorization request",
					})
				}

				if tokenErr.Is(xbox.BadVerificationCodeErr) {
					return locales.Translate(&i18n.Message{
						ID:    "command.accounts-add-microsoft.error.bad-verification-code",
						Other: "Cannot check authorization status: the device code is not recognized by Microsoft",
					})
				}
			}

			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": stepErr.Err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.device-auth-failed",
					Other: "Cannot authorize the device: {{ .Error }}",
				},
			})

		case xboxAuthFlow.StepIDXBLAuth:
			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": stepErr.Err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.xbl-auth-failed",
					Other: "Failed to log you into Xbox Live: {{ .Error }}",
				},
			})

		case xboxAuthFlow.StepIDXSTSAuth:
			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": stepErr.Err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.xsts-token-request-failed",
					Other: "Cannot acquire Xbox Secure Token: {{ .Error }}",
				},
			})

		case xboxAuthFlow.StepIDMCAuth:
			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.minecraft-login-failed",
					Other: "Cannot log you in: {{ .Error }}",
				},
			})

		case xboxAuthFlow.StepIDEntitlementsCheck:
			if errors.Is(stepErr.Err, &xboxAuthFlow.EntitlementMissingError{}) {
				return locales.Translate(&i18n.Message{
					ID: "command.accounts-add-microsoft.error.not-owns-game",
					Other: "You do not appear to own Minecraft.\n" +
						"Buy Minecraft at https://www.minecraft.net/store/minecraft-java-edition\n" +
						"If you are using Xbox PC Game Pass, you need to log in into official launcher at least once.",
				})
			}

			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": stepErr.Err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.entitlement-check-failed",
					Other: "Cannot make sure that you own Minecraft: {{ .Error }}",
				},
			})

		case xboxAuthFlow.StepIDRefreshMCProfile:
			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": stepErr.Err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.profile-refresh-failed",
					Other: "Cannot fetch your Minecraft profile: {{ .Error }}",
				},
			})

		case xboxAuthFlow.StepIDFlushData:
			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": stepErr.Err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-microsoft.error.data-flush-failed",
					Other: "Cannot save secure account authentication data",
				},
			})
		}
	}

	return locales.TranslateUsing(&i18n.LocalizeConfig{
		TemplateData: map[string]string{
			"Error": err.Error(),
		},
		DefaultMessage: &i18n.Message{
			ID:    "command.accounts-add-microsoft.error.unknown-error",
			Other: "Cannot authorize you: unknown error; {{ .Error }}",
		},
	})
}
