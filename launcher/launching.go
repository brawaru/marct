package launcher

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/brawaru/marct/launcher/accounts"
	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/utils/slices"
	"github.com/brawaru/marct/validfile"
	"github.com/google/shlex"
)

const DefaultJVMArgs = "-Xmx2G -XX:+UnlockExperimentalVMOptions -XX:+UseG1GC -XX:G1NewSizePercent=20 -XX:G1ReservePercent=20 -XX:MaxGCPauseMillis=50 -XX:G1HeapRegionSize=32M"

// standardJVMArguments returns standard Java arguments that Minecraft launcher uses when client JSON does not have them specified.
func standardJVMArguments() (args []Argument) {
	b, err := resourcesFS.ReadFile("resources/jvm_args.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, &args)
	if err != nil {
		panic(err)
	}
	return
}

type LaunchOptions struct {
	Background    bool                   // Whether to launch the game in background with attacking it a child.
	JavaPath      string                 // Path to javaw executable.
	JavaArgs      *string                // User JVM arguments.
	Resolution    *Resolution            // Custom resolution.
	Authorization accounts.Authorization // Authorization.
	GameDirectory string                 // Directory where game stores its files like resource packs.
}

type LaunchResult struct {
	Command           *exec.Cmd
	NativesDirectory  string
	VirtualAssetsPath string
}

type CleanError struct {
	Suppressed []error
}

func (e *CleanError) Error() string {
	s := "failed to clean temporary directories"
	for _, i := range e.Suppressed {
		s += "\n  suppressed: " + i.Error()
	}
	return s
}

func (r *LaunchResult) Clean() error {
	var s []error

	if err := os.RemoveAll(r.NativesDirectory); err != nil {
		s = append(s, err)
	}

	if err := os.RemoveAll(r.VirtualAssetsPath); err != nil {
		s = append(s, err)
	}

	if len(s) == 0 {
		return nil
	} else {
		return &CleanError{
			Suppressed: s,
		}
	}
}

// isBannedArgument checks whether argument must be skipped since it cannot be provided by launcher and not important.
func isBannedArgument(v string) bool {
	return v == "${clientid}" || v == "${auth_xuid}" || v == "--clientId" || v == "--xuid"
}

func interpretArgv(argv []string, vars map[string]string) []string {
	r := regexp.MustCompile(`(?i)\$\{(?P<Name>[A-Z_\-]+)}`)
	c := slices.Copy(argv)

	for i, a := range c { // FIXME: this should apply to intermediate argv that doesn't have things like mainClass
		for {
			m := utils.MapRegexMatches(r, a)
			if m == nil {
				break
			}
			v := vars[m["Name"]]
			a = strings.ReplaceAll(a, m[""], v)
		}

		c[i] = a
	}

	return c
}

func (w *Instance) Launch(version Version, options LaunchOptions) (*LaunchResult, error) {
	lr := &LaunchResult{}

	// Minecraft arguments:
	// - ${auth_player_name}   Name of the player.
	// - ${version_name}       Identifier of the version launched.
	// - ${game_directory}     Absolute path to the game directory.
	// - ${assets_root}        Absolute path to the assets directory.
	// - ${assets_index_name}  Asset index name.
	// - ${auth_uuid}          User UUID.
	// - ${auth_access_token}  Access token.
	// - ${user_type}          User type (msa or mojang).
	// - ${version_type}       Version type (release, snapshot, etc).
	// - ${resolution_width}   Resolution width.
	// - ${resolution_height}  Resolution height.
	// == DO NOT IMPLEMENT ON CUSTOM CLIENT: ==
	// - ${clientid}           Client ID (analytics).
	// - ${auth_xuid}          XUID of the player (analytics).
	//
	// JVM arguments:
	// - ${natives_directory}  Absolute path to virtualised natives.
	// - ${launcher_name}      Name of the launcher.
	// - ${launcher_version}   Launcher version.
	// - ${classpath}          Classpath (absolute paths to all libraries separated by `;`).

	// minecraftArguments placeholders:
	// ${auth_player_name} ${auth_session} --gameDir ${game_directory} --assetsDir ${game_assets} --tweakClass net.minecraftforge.legacy._1_5_2.LibraryFixerTweaker
	//
	// - ${auth_player_name}  Name of the user.
	// - ${auth_session}      Auth session in format (often token:[minecraft token]).
	// - ${game_directory}    Absolute path to the game directory.
	// - ${game_assets}       Path to virtualised assets.
	// - ${width}             Custom resolution width.
	// - ${height}            Custom resolution height.

	gameDirectory := filepath.FromSlash(options.GameDirectory)

	if gameDirectory == "" || !filepath.IsAbs(gameDirectory) {
		gameDirectory = filepath.Join(w.Path, gameDirectory)
	}

	if dirExists, err := validfile.DirExists(gameDirectory); err == nil {
		if !dirExists {
			if err := os.MkdirAll(gameDirectory, 0755); err != nil {
				return nil, fmt.Errorf("cannot create game directory: %w", err)
			}
		}
	} else {
		return nil, fmt.Errorf("cannot check game directory: %w", err)
	}

	var virtualAssetsPath string
	if version.AssetIndex != nil {
		if ai, err := w.ReadAssetIndex(version.AssetIndex.ID); err != nil {
			return lr, err
		} else {
			var vp string
			switch ai.MapType() {
			case AsResources:
				vp = filepath.Join(gameDirectory, "resources")
			case AsVirtual:
				vp = w.AssetsVirtualPath(version.AssetIndex.ID)
			}

			if vp != "" {
				// FIXME: assets that have to be mapped in virtual directory can be hardlinked to save space and time on copying
				//  since they are always in the same place unlike assets that are mapped as resources.
				if err := ai.Virtualize(w.DefaultAssetsObjectResolver(), vp); err != nil {
					return nil, fmt.Errorf("cannot virtualize assets: %w", err)
				}

				virtualAssetsPath = vp
				lr.VirtualAssetsPath = vp // FIXME: assets mapped as objects shall not be removed since this is a costly operation
			}
		}
	}

	var nativesDirectory string
	if np, err := w.ExtractNatives(version); err == nil {
		nativesDirectory = np
		lr.NativesDirectory = np
	} else {
		return nil, fmt.Errorf("extract natives: %w", err)
	}

	ld := w.LibrariesPath()
	var classPath []string
	for _, l := range version.Libraries {
		if l.Rules != nil && !l.Rules.Matches() {
			continue
		}
		lp := filepath.Join(ld, l.Coordinates.Path(os.PathSeparator))
		if !strings.HasPrefix(lp, ld+string(os.PathSeparator)) {
			return nil, fmt.Errorf("illegal library path: %q", lp)
		}

		classPath = append(classPath, lp)
	}

	path, err := w.VersionFilePath(version.ID, "jar")
	if err != nil {
		return nil, fmt.Errorf("path %q to version jar: %w", version.ID, err)
	}
	classPath = append(classPath, path)

	var resWidth string
	var resHeight string
	{
		r := options.Resolution.Max(Resolution{
			Width:  300,
			Height: 260,
		})

		resWidth = strconv.Itoa(r.Width)
		resHeight = strconv.Itoa(r.Height)
	}

	featSet := map[Feature]bool{
		FeatDemoUser:         options.Authorization.DemoUser,
		FeatCustomResolution: options.Resolution != nil,
	}

	var jvmArgv []string

	{
		var jvmArguments []Argument
		if version.Arguments != nil {
			jvmArguments = version.Arguments.JVM
		} else {
			jvmArguments = standardJVMArguments()
		}

		for _, argument := range jvmArguments {
			if argument.Rules.MatchesExtensively(featSet) {
				for _, s := range argument.Value {
					if isBannedArgument(s) {
						continue
					}
					jvmArgv = append(jvmArgv, s)
				}
			}
		}
	}

	var minecraftArgv []string

	if version.Arguments != nil && version.Arguments.Game != nil {
		for _, argument := range version.Arguments.Game {
			if argument.Rules.MatchesExtensively(featSet) {
				for _, a := range argument.Value {
					if !isBannedArgument(a) {
						minecraftArgv = append(minecraftArgv, a)
					}
				}
			}
		}
	} else if version.MinecraftArguments != nil {
		s := *version.MinecraftArguments
		for _, a := range strings.Split(s, " ") {
			if !isBannedArgument(a) {
				minecraftArgv = append(minecraftArgv, a)
			}
		}
	}

	argvVars := map[string]string{
		"auth_player_name":  options.Authorization.UserName,
		"version_name":      version.ID,
		"game_directory":    gameDirectory,
		"assets_root":       filepath.Join(w.Path, assetsPath),
		"assets_index_name": *version.Assets,
		"auth_uuid":         options.Authorization.UserUUID,
		"auth_access_token": options.Authorization.AccessToken,
		"user_type":         options.Authorization.UserType,
		"version_type":      *version.Type,
		"resolution_width":  resWidth,
		"resolution_height": resHeight,
		"natives_directory": nativesDirectory,
		"launcher_name":     "Marct",
		"launcher_version":  "1.0.0",
		"classpath":         strings.Join(classPath, ";"),
		"auth_session":      "token:" + options.Authorization.AccessToken,
		"game_assets":       virtualAssetsPath, // FIXME: only for virtual directories
		"width":             resWidth,
		"height":            resHeight,
	}

	var loggingArgv []string
	var loggingProcessor io.Writer

	if version.Logging != nil {
		c, ok := version.Logging["client"]
		if ok {
			loggingArgv = interpretArgv([]string{c.Argument}, map[string]string{
				"path": w.LogConfigPath(c),
			})

			switch c.Type {
			case "log4j2-xml":
				loggingProcessor = &Log4JWriter{
					Consumer: func(event Log4JEvent) {
						fmt.Printf("[%s] [%s] [%s] %s\n", event.Timestamp.Format(time.RFC3339), event.Level, event.Thread, event.Message.Content)
					},
				}
			}
		}
	}

	if loggingProcessor == nil {
		loggingProcessor = os.Stdout
	}

	var userArgv []string

	{
		var s string

		if options.JavaArgs == nil {
			s = DefaultJVMArgs
		} else {
			s = *options.JavaArgs
		}

		a, err := shlex.Split(s)

		if err != nil {
			return nil, fmt.Errorf("cannot parse JVM arguments %q: %w", *options.JavaArgs, err)
		}

		userArgv = a
	}

	// [JVM arguments] [user JVM arguments] [config argument] [mainClass] [minecraft arguments]

	var argv []string
	argv = append(argv, interpretArgv(jvmArgv, argvVars)...)
	argv = append(argv, userArgv...)
	argv = append(argv, loggingArgv...)
	argv = append(argv, version.MainClass)
	argv = append(argv, interpretArgv(minecraftArgv, argvVars)...)

	var javawPath string

	if options.JavaPath == "" {
		var j string
		if version.JavaVersion == nil {
			j = "jre-legacy"
		} else {
			j = version.JavaVersion.Component
		}
		javawPath = filepath.FromSlash(fmt.Sprintf("%s/%s/bin/java", w.JREPath(j, GetJRESelector()), j))
		if runtime.GOOS == "windows" {
			javawPath += ".exe"
		}
	} else {
		javawPath = options.JavaPath
	}

	fmt.Printf("java: %q\nargv:\n %s\n", javawPath, strings.Join(argv, "\n "))

	cmd := exec.Command(javawPath, argv...)
	cmd.Env = os.Environ()
	cmd.Dir = gameDirectory
	cmd.Stdout = loggingProcessor
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	lr.Command = cmd

	return lr, nil
}
