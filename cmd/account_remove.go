package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	minecraftAccount "github.com/brawaru/marct/minecraft/account"
	"github.com/brawaru/marct/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var accountRemoveCommand = createCommand(&cli.Command{
	Name:    "remove",
	Aliases: []string{"rm"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.accounts-remove.usage",
		Other: "Remove an account",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.accounts-remove.description",
		Other: "Remove an existing account",
	}),
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name: "force",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.accounts-remove.flag.force.usage",
				Other: "Do not ask questions",
			}),
		},
	},
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.accounts-remove.args-usage",
		Other: "<account identifier>",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value("workDir").(*launcher.Instance)

		if ctx.NArg() < 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.accounts-remove.error.no-account-identifier",
				Other: "No account identifier provided",
			}), 1)
		} else if ctx.NArg() > 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.accounts-remove.error.too-many-arguments",
				Other: "Too many arguments provided",
			}), 1)
		}

		accountID := ctx.Args().First()

		store, err := workDir.OpenAccountsStore()

		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-remove.error.failed-to-open-accounts-store",
					Other: "Failed to open accounts store: {{.Error}}",
				},
			}), 1)
		}

		defer utils.DClose(store)

		account, ok := store.Accounts[accountID]

		if !ok {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.accounts-remove.error.account-not-found",
				Other: "Account not found",
			}), 1)
		}

		if !ctx.Bool("force") {
			var answer bool

			properties, err := minecraftAccount.ReadProperties(&account)

			if err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.accounts-remove.error.failed-to-read-properties",
						Other: "Failed to read account properties: {{.Error}}",
					},
				}), 1)
			}

			err = survey.AskOne(&survey.Confirm{
				Message: locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"AccountName": properties.Username,
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.accounts-remove.confirm.message",
						Other: "Are you sure you want to remove account {{.AccountID}}?",
					},
				}),
				Default: false,
			}, &answer)

			if err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": err.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.accounts-remove.error.failed-to-ask-confirmation",
						Other: "Failed to ask confirmation: {{.Error}}",
					},
				}), 1)
			}

			if !answer {
				return nil
			}
		}

		delete(store.Accounts, accountID)

		if err := store.Save(); err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-remove.error.failed-to-save-accounts-store",
					Other: "Failed to save changes: {{.Error}}",
				},
			}), 1)
		}

		return nil
	},
})

func init() {
	accountCommand.Subcommands = append(accountCommand.Subcommands, accountRemoveCommand)
}
