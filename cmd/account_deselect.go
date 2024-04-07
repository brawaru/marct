package cmd

import (
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var accountDeselectCommand = createCommand(&cli.Command{
	Name:    "deselect",
	Aliases: []string{"clsel"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.accounts-deselect.usage",
		Other: "Clear default account selection",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.accounts-deselect.description",
		Other: "Clears previously set default account seleciton",
	}),
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value(instanceKey).(*launcher.Instance)

		accountsStore, err := workDir.OpenAccountsStore()
		if err != nil {
			return cli.Exit(locales.TranslateWith(&i18n.Message{
				ID:    "command.accounts-deselect.error.cannot-open-store",
				Other: "Cannot open accounts store: {{ .Error }}",
			}, map[string]string{
				"Error": err.Error(),
			}), 1)
		}

		defer utils.DClose(accountsStore)

		accountsStore.SelectedAccount = ""

		if err := accountsStore.Save(); err != nil {
			return cli.Exit(locales.TranslateWith(&i18n.Message{
				ID:    "command.accounts-deselect.error.cannot-save-store",
				Other: "Cannot save store: {{ .Error }}",
			}, map[string]string{
				"Error": err.Error(),
			}), 1)
		}

		println(locales.Translate(&i18n.Message{
			ID:    "command.accounts-deselect.error.done",
			Other: "Account selection cleared",
		}))

		return nil
	},
})
