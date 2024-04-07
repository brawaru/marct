package cmd

import (
	"context"
	"errors"

	"github.com/99designs/keyring"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/locales"
	offlineAuthFlow "github.com/brawaru/marct/offline/authflow"
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
	Subcommands: []*cli.Command{accountAddMicrosoftCommand, accountAddOfflineCommand},
	Before: func(ctx *cli.Context) error {
		instance := ctx.Context.Value(instanceKey).(*launcher.Instance)

		store, err := instance.OpenAccountsStore()

		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add.open-store-error",
					Other: "Cannot open accounts store: {{.Error}}",
				},
			}), 1)
		}

		ctx.Context = context.WithValue(ctx.Context, accountsStoreKey, store)

		return nil
	},

	After: func(ctx *cli.Context) error {
		store := ctx.Context.Value(accountsStoreKey).(*accounts.StoreFile)
		err := store.Close()

		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add.close-store-error",
					Other: "Cannot close accounts store: {{.Error}}",
				},
			}), 1)
		}

		return nil
	},
})

func init() {
	accountCommand.Subcommands = append(accountCommand.Subcommands, accountAddCommand)
}

var accountAddOfflineCommand = createCommand(&cli.Command{
	Name: "offline",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.accounts-add-offline.usage",
		Other: "A new offline account",
	}),
	Description: locales.Translate(&i18n.Message{
		ID: "command.accounts-add-offline.description",
		Other: "Allows to set up a new offline account that works even when no connection is available.\n" +
			"This account, however, will not be usable to connect to the online-mode servers.",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.accounts-add-offline.args-usage",
		Other: "[username]",
	}),
	Action: func(ctx *cli.Context) error {
		store := ctx.Context.Value(accountsStoreKey).(*accounts.StoreFile)

		if ctx.NArg() > 1 {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-offline.too-many-args",
					Other: "Too many arguments, expected only one username",
				},
			}), 1)
		}

		authFlow := offlineAuthFlow.CreateAuthFlow(&offlineAuthFlow.Options{
			UsernameRequestHandler: func() (string, error) {
				var username string

				if ctx.NArg() == 0 {
					return offlineUsernamePrompt()
				}

				username = ctx.Args().Get(0)

				if !minecraftUsernameRegex.MatchString(username) {
					return "", errors.New(locales.Translate(&i18n.Message{
						ID:    "command.accounts-add-offline.invalid-username",
						Other: "Invalid username supplied as an argument",
					}))
				}

				return username, nil
			},
		})

		account, err := authFlow.CreateAccount()

		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-add-offline.create-account-error",
					Other: "Cannot create account: {{.Error}}",
				},
			}), 1)
		}

		store.AddAccount(account)

		return nil
	},
})

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
		instance := ctx.Context.Value(instanceKey).(*launcher.Instance)

		k, err := keyringOpenFlow(instance)
		if err != nil {
			return err
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

		accountsStore := ctx.Context.Value(accountsStoreKey).(*accounts.StoreFile)

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

		accountsStore.Store.Accounts[account.ID] = &*&account

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
