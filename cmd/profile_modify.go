package cmd

import (
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var profileModifyCommand = createCommand(&cli.Command{
	Name:    "modify",
	Aliases: []string{"edit"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.profile-modify.usage",
		Other: "Modify existing profile settings",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.profile-modify.description",
		Other: "Modifies an existing profile using the provided flag values",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.profile-modify.args-usage",
		Other: "[identifier]",
	}),
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "name",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-create.args.name",
				Other: "Display name of the profile",
			}),
		},
		&cli.StringFlag{
			Name: "version",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-create.args.version",
				Other: "Game version that the profile uses",
			}),
		},
		&cli.StringFlag{
			Name: "icon",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-create.args.icon",
				Other: "Profile icon (Minecraft Launcher)",
			}),
		},
		&cli.PathFlag{
			Name: "path",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-create.args.path",
				Other: "Game files path",
			}),
		},
		&cli.StringSliceFlag{
			Name: "jvm-args",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-create.args.jvm-args",
				Other: "JVM arguments",
			}),
		},
		&cli.StringFlag{
			Name: "java-path",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-create.args.java-path",
				Other: "Java executable path",
			}),
		},
		&cli.StringFlag{
			Name: "resolution",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-create.args.resolution",
				Other: "Window resolution (e.g. 1280x720)",
			}),
			DefaultText: "auto",
		},
	},
	Action: func(ctx *cli.Context) error {
		profileID := ctx.Args().First()

		if len(profileID) == 0 {
			// ask if interactive?
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.profile-modify.empty-id",
				Other: "You must provide of the profile you are willing to modify",
			}), 1)
		}

		profiles := ctx.Context.Value(profilesKey).(*launcher.Profiles)

		profile, profileExists := profiles.Profiles[profileID]

		if !profileExists {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"ID": profileID,
				},
				DefaultMessage: nil,
			}), 1)
		}

		// name, version, icon, path, jvm-args, java-path, resolution

		if ctx.IsSet("name") {
			profile.Name = ctx.String("name")
		}

		if ctx.IsSet("version") {
			profile.LastVersionID = ctx.String("version")
		}

		if ctx.IsSet("icon") {
			// FIXME: there must be a check if icon is valid or not
			icon := ctx.String("icon")
			profile.Icon = &icon
		}

		if ctx.IsSet("path") {
			profile.GameDir = ctx.String("path")
		}

		if ctx.IsSet("jvm-args") {
			jvmArgs := ctx.String("jvm-args")
			profile.JavaArgs = &jvmArgs
		}

		if ctx.IsSet("java-path") {
			jvmPath := ctx.String("java-path")
			profile.JavaPath = &jvmPath
		}

		if ctx.IsSet("resolution") {
			resolution, parseErr := launcher.ParseResolution(ctx.String("resolution"))
			if parseErr != nil {
				return cli.Exit(translateResolutionParseError(parseErr), 1)
			}
			profile.Resolution = resolution
		}

		return nil
	},
})

func init() {
	app.Commands = append(app.Commands, profileModifyCommand)
}
