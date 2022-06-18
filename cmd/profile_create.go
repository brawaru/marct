package cmd

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/brawaru/marct/launcher"
	"github.com/brawaru/marct/locales"
	"github.com/brawaru/marct/sdtypes"
	"github.com/brawaru/marct/utils"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/urfave/cli/v2"
)

var msgSurveyFail = &i18n.Message{
	ID:    "command.profile-create.error.survey-fail",
	Other: "Cannot acquire value for parameter '{{ .Name }}'",
}

var profileCreateCommand = createCommand(&cli.Command{
	Name:    "create",
	Aliases: []string{"new"},
	Usage: locales.Translate(&i18n.Message{
		ID:    "command.profile-create.usage",
		Other: "Create new profile",
	}),
	Description: locales.Translate(&i18n.Message{
		ID:    "command.profile-create.description",
		Other: "Create a new profile using provided flag values or by answering interactive questions",
	}),
	ArgsUsage: locales.Translate(&i18n.Message{
		ID:    "command.profile-create.args-usage",
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
			Name:    "java-args",
			Aliases: []string{"jvm-args"},
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
		&cli.BoolFlag{
			Name: "overwrite",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-create.args.overwrite",
				Other: "Overwrite existing profile with the same ID",
			}),
		},
		&cli.BoolFlag{
			Name: "defaults",
			Usage: locales.Translate(&i18n.Message{
				ID:    "command.profile-create.args.defaults",
				Other: "Use defaults instead of asking",
			}),
		},
	},
	Action: func(ctx *cli.Context) error {
		workDir := ctx.Context.Value(instanceKey).(*launcher.Instance)
		profiles := ctx.Context.Value(profilesKey).(*launcher.Profiles)

		if ctx.NArg() > 1 {
			return cli.Exit(locales.Translate(&i18n.Message{
				ID:    "command.profile-create.error.too-many-args",
				Other: "Too many arguments: expected only profile ID",
			}), 1)
		}

		id := ctx.Args().First()

		if len(id) == 0 {
			id = utils.NewUUID()
		}

		overwrite := ctx.Bool("overwrite")

		if !overwrite {
			if profiles.Profiles != nil {
				_, exists := profiles.Profiles[id]

				if exists {
					return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
						TemplateData: map[string]string{
							"ID": id,
						},
						DefaultMessage: &i18n.Message{
							ID:    "command.profile-create.error.id-exists",
							Other: "Profile with ID \"{{ .ID }}\" already exists. Use --overwrite flag to overwrite it.",
						},
					}), 1)
				}
			}
		}

		defaults := ctx.Bool("defaults")

		var profile launcher.Profile

		creationTime := sdtypes.ISOTime(time.Now())

		var icon = "Furnace"

		profile = launcher.Profile{
			Created:       &creationTime,
			Icon:          &icon,
			LastUsed:      sdtypes.ISOTime{},
			LastVersionID: "",
			Name:          "",
			Type:          "custom",
			JavaArgs:      nil,
			JavaPath:      nil,
			Resolution:    nil,
			GameDir:       "",
		}

		if ctx.IsSet("name") {
			profile.Name = ctx.String("name")
		} else if defaults {
			profile.Name = id
		} else {
			surveyErr := survey.AskOne(&survey.Input{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.profile-create.survey.name",
					Other: "Name of the profile",
				}),
				Default: id,
			}, &profile.Name)

			if surveyErr != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Name": "name",
					},
					DefaultMessage: msgSurveyFail,
				}), 1)
			}
		}

		if ctx.IsSet("version") {
			profile.LastVersionID = ctx.String("version")
		} else if defaults {
			profile.LastVersionID = "latest-release"
		} else {
			_, _ = os.Stdout.WriteString(locales.Translate(&i18n.Message{
				ID:    "command.profile-create.fetching-versions",
				Other: "Fetching versions, please wait...",
			}) + "\r")

			manifest, fetchErr := workDir.FetchVersions(false)
			if fetchErr != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Error": fetchErr.Error(),
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.profile-create.versions",
						Other: "Failed to versions due to error: {{ .Error }}",
					},
				}), 1)
			}

			var options []string
			optionsMappings := map[string]string{}

			{
				var latestReleaseText = locales.Translate(&i18n.Message{
					ID:    "command.profile-create.survey.version.option.latest-release",
					Other: "Latest release",
				})

				options = append(options, latestReleaseText)
				optionsMappings[latestReleaseText] = "latest-release"
			}

			{
				var latestSnapshotText = locales.Translate(&i18n.Message{
					ID:    "command.profile-create.survey.version.option.latest-snapshot",
					Other: "Latest snapshot",
				})

				options = append(options, latestSnapshotText)
				optionsMappings[latestSnapshotText] = "latest-snapshot"
			}

			for _, version := range manifest.Versions {
				options = append(options, version.ID)
				optionsMappings[version.ID] = version.ID
			}

			var selectedOption string

			surveyErr := survey.AskOne(&survey.Select{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.profile-create.survey.version",
					Other: "Version to use",
				}),
				Options: options,
			}, &selectedOption)

			if surveyErr != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Name": "version",
					},
					DefaultMessage: msgSurveyFail,
				}), 1)
			}

			version, exists := optionsMappings[selectedOption]

			if !exists {
				return cli.Exit(locales.Translate(&i18n.Message{
					ID:    "command.profile-create.error.survey-fail-version",
					Other: "Illegal option selected for 'version'",
				}), 1)
			}

			profile.LastVersionID = version
		}

		if ctx.IsSet("icon") {
			icon = ctx.String("icon")
		} else if defaults {
			icon = "Furnace"
		} else {
			surveyErr := survey.AskOne(&survey.Select{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.profile-create.survey.icon",
					Other: "Select icon",
				}),
				Options: launcher.LauncherIcons,
				Default: "Furnace",
			}, &icon)

			if surveyErr != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Name": "icon",
					},
					DefaultMessage: msgSurveyFail,
				}), 1)
			}
		}

		if ctx.IsSet("path") {
			profile.GameDir = ctx.Path("path")
		} else if defaults {
			profile.GameDir = ""
		} else {
			surveyErr := survey.AskOne(&survey.Input{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.profile-create.survey.path",
					Other: "Game directory",
				}),
				// TODO: add path suggestion
			}, &profile.GameDir)

			if surveyErr != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Name": "path",
					},
					DefaultMessage: msgSurveyFail,
				}), 1)
			}
		}

		if ctx.IsSet("resolution") {
			resolution, parseErr := launcher.ParseResolution(ctx.String("resolution"))

			if parseErr != nil {
				return cli.Exit(translateResolutionParseError(parseErr), 1)
			}

			profile.Resolution = resolution
		} else if defaults {
			profile.Resolution = nil
		} else {
			defaultResolutions := []string{
				"auto",
				"800x600",
				"1024x768",
				"1280x720",
				"1280x1024",
				"1366x768",
				"1536x864",
				"1440x900",
				"1600x900",
				"1920x1080",
				"2560x1440",
			}

			var resp string

			surveyErr := survey.AskOne(&survey.Input{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.profile-create.survey.screen-resolution",
					Other: "Screen resolution",
				}),
				Suggest: func(toComplete string) (res []string) {
					for _, resolution := range defaultResolutions {
						if strings.HasPrefix(resolution, toComplete) {
							res = append(res, resolution)
						}
					}

					return
				},
				Default: "auto",
			}, &resp, survey.WithValidator(func(ans interface{}) error {
				s, ok := ans.(string)

				if !ok {
					return errors.New(locales.Translate(&i18n.Message{
						ID:    "command.profile-create.error.invalid-type",
						Other: "Expected answer of string type",
					}))
				}

				if _, parseErr := launcher.ParseResolution(s); parseErr != nil {
					return errors.New(translateResolutionParseError(parseErr))
				}

				return nil
			}))

			if surveyErr != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Name": "resolution",
					},
					DefaultMessage: msgSurveyFail,
				}), 1)
			}

			if resolution, parseErr := launcher.ParseResolution(resp); parseErr == nil {
				profile.Resolution = resolution
			} else {
				return cli.Exit(translateResolutionParseError(parseErr), 1)
			}
		}

		if ctx.IsSet("java-args") {
			s := ctx.String("java-args")
			profile.JavaArgs = &s
		} else if defaults {
			profile.JavaArgs = nil
		} else {
			var s string

			if err := survey.AskOne(&survey.Input{
				Message: locales.Translate(&i18n.Message{
					ID:    "command.profile-create.survey.jvm-args",
					Other: "JVM arguments",
				}),
				Help: locales.Translate(&i18n.Message{
					ID:    "command.profile-create.survey.jvm-args.help",
					Other: "Arguments for Java Virtual Machine. Use default value if you don't know what these mean.",
				}),
				Default: launcher.DefaultJVMArgs,
			}, &s); err != nil {
				return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Name": "java-args",
					},
					DefaultMessage: msgSurveyFail,
				}), 1)
			}

			if s == launcher.DefaultJVMArgs {
				profile.JavaArgs = nil
			} else {
				profile.JavaArgs = &s
			}
		}

		if profiles.Profiles == nil {
			profiles.Profiles = map[string]launcher.Profile{}
		}

		profiles.Profiles[id] = profile

		if writeErr := workDir.WriteProfiles(profiles); writeErr != nil {
			return cli.Exit(locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Error": writeErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.profile-create.error.profiles-write-error",
					Other: "Failed to write profiles file: {{ .Error }}",
				},
			}), 1)
		}

		return nil
	},
})

func translateResolutionParseError(err error) string {
	{
		var e *launcher.IllegalDimensionsNumberError
		if errors.As(err, &e) {
			if e.Count == -1 {
				return locales.Translate(&i18n.Message{
					ID:    "command.profile-create.error.illegal-dimensions",
					Other: "No dimensions provided",
				})
			}

			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Count": strconv.Itoa(e.Count),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.profile-create.error.illegal-dimensions",
					Other: "Illegal number of dimensions provided - {{ .Count }}",
				},
			})
		}
	}

	{
		var e *launcher.IllegalDimensionValueError
		if errors.As(err, &e) {
			subErr := errors.Unwrap(e)

			var dimension string
			switch e.Dimension {
			case "width":
				dimension = locales.Translate(&i18n.Message{
					ID:    "command.profile-create.dimension-width",
					Other: "width",
				})
			case "height":
				dimension = locales.Translate(&i18n.Message{
					ID:    "command.profile-create.dimension-height",
					Other: "height",
				})
			}

			if subErr == nil {
				return locales.TranslateUsing(&i18n.LocalizeConfig{
					TemplateData: map[string]string{
						"Dimension": dimension,
						"Value":     e.Value,
					},
					DefaultMessage: &i18n.Message{
						ID:    "command.profile-create.error.illegal-dimension-value",
						Other: "Invalid value for dimension {{ .Dimension }} - {{ .Value }}",
					},
				})
			}

			return locales.TranslateUsing(&i18n.LocalizeConfig{
				TemplateData: map[string]string{
					"Dimension": dimension,
					"Value":     e.Value,
					"Error":     subErr.Error(),
				},
				DefaultMessage: &i18n.Message{
					ID:    "command.profile-create.error.illegal-dimension-value-error",
					Other: "Invalid value for dimension {{ .Dimension }} - {{ .Value }}: {{ .Error }}",
				},
			})
		}
	}

	return locales.TranslateUsing(&i18n.LocalizeConfig{
		TemplateData: map[string]string{
			"Error": err.Error(),
		},
		DefaultMessage: &i18n.Message{
			ID:    "command.profile-create.error.resolution-parse-error",
			Other: "Cannot read resolution: {{ .Error }}",
		},
	})
}

func init() {
	profileCommand.Subcommands = append(profileCommand.Subcommands, profileCreateCommand)
}
