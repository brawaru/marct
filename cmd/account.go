package cmd

import (
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var accountCommand = createCommand(&cli.Command{
	Name:    "account",
	Aliases: []string{"accounts", "acc", "accs"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.accounts.usage",
		Other: "Manage accounts used to log in to game",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.accounts.description",
		Other: "This command allows to manage accounts used to log in to game",
	}),
	Subcommands: nil,
})

func init() {
	app.Commands = append(app.Commands, accountCommand)
}
