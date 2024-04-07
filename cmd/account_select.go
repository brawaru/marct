package cmd

import (
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/minecraft/account"
	"github.com/brawaru/marct/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var accountSelectCommand = createCommand(&cli.Command{
	Name:    "select",
	Aliases: []string{"sel"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.accounts-select.usage",
		Other: "Select default account",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.accounts-select.description",
		Other: "Allows to select a default account that will be used when launching the game to skip the account selection prompt",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.accounts-select.args-usage",
		Other: "<account identifier>",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value(instanceKey).(*launcher.Instance)

		if ctx.NArg() > 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.accounts-select.error.too-many-arguments",
				Other: "Too many arguments provided",
			}), 1)
		}

		accountsStore, err := workDir.OpenAccountsStore()
		if err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-select.error.accounts-store-open-error",
					Other: "Cannot open accounts store: {{ .Error }}",
				},
			}), 1)
		}

		defer utils.DClose(accountsStore)

		var selectedAccount *accounts.Account

		if ctx.NArg() > 0 {
			lookupId := ctx.Args().First()

			for _, account := range accountsStore.Accounts {
				if account.ID == lookupId {
					selectedAccount = account
					break
				}
			}

			if selectedAccount == nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"ID": lookupId,
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.accounts-select.error.no-account-with-id",
						Other: "No account with ID: {{ .ID }}",
					},
				}), 1)
			}
		} else {
			if selectedAccount, err = SelectAccountFlow(accountsStore.Accounts); err != nil {
				return err
			}
		}

		accountsStore.SelectedAccount = selectedAccount.ID

		if err := accountsStore.Save(); err != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": err.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-select.error.cannot-save-store",
					Other: "Cannot save store due to error: {{ .Error }}",
				},
			}), 1)
		}

		if minecraftProperties, err := account.ReadProperties(selectedAccount); err == nil {
			println(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Username": minecraftProperties.Username,
					"Type":     selectedAccount.Type,
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-select.selected",
					Other: "{{ .Username }} ({{ .Type }}) is selected as default (✿◡‿◡)",
				},
			}))
		} else {
			println(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"ID":   selectedAccount.ID,
					"Type": selectedAccount.Type,
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.accounts-select.selected-alternative",
					Other: "{{ .ID }} ({{ .Type }}) is selected as default (✿◡‿◡)",
				},
			}))
		}

		return nil
	},
})

func init() {
	accountCommand.Subcommands = append(accountCommand.Subcommands, accountSelectCommand)
}
