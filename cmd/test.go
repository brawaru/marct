package cmd

import (
	"fmt"

	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var testCmd = createCommand(&cli.Command{
	Name: "test",
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.test.usage",
		Other: "Test",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.test.description",
		Other: "This command is used for internal testing",
	}),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "ewe",
			Usage: "ewe val",
		},
	},
	Action: func(ctx *cli.Context) error {
		first := ctx.Args().First()
		tail := ctx.Args().Tail()
		sli := ctx.Args().Slice()
		fmt.Printf("%v, %#v, %#v", first, tail, sli)
		return nil
	},
})

func init() {
	app.Commands = append(app.Commands, testCmd)
}
