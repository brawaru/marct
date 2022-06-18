package cmd

import (
	"context"

	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

const profilesKey ctxKey = "profiles"

var profileCommand = createCommand(&cli.Command{
	Name:    "profile",
	Aliases: []string{"profiles"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.profile.usage",
		Other: "Manage game profiles",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.profile.description",
		Other: "This command allows you to manage game profiles",
	}),
	Before: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value(instanceKey).(*launcher.Instance)

		profiles, err := readProfilesOrExit(workDir)

		if err != nil {
			return err
		}

		ctx.Context = context.WithValue(ctx.Context, profilesKey, profiles)

		return nil
	},
})

func readProfilesOrExit(w *launcher.Instance) (*launcher.Profiles, error) {
	profiles, err := w.ReadOrCreateProfiles()
	if err != nil {
		return nil, cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
			TemplateData: map[string]string{
				"Error": err.Error(),
			},
			DefaultMessage: &i18n.Message{
				ID:    "command.profile-create.error.profiles-read",
				Other: "Cannot read profiles file: {{ .Error }}",
			},
		}), 1)
	}
	return profiles, nil
}

func init() {
	app.Commands = append(app.Commands, profileCommand)
}
