package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/brawaru/marct/cmd"
	"github.com/google/shlex"
)

//go:generate goi18n extract -sourceLanguage en-US -outdir locales

func main() {
	var argv []string
	if v, ok := os.LookupEnv("TERM_PROGRAM"); ok && v == "vscode" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fmt.Printf("PWD: %v\n", wd)

		print("Arguments: ")
		// read argv from stdin
		b, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			panic(err)
		}

		// remove trailing \r\n
		b = strings.TrimRight(b, "\r\n")

		// parse b as argv
		i, err := shlex.Split(b)
		if err != nil {
			panic(err)
		}

		if len(os.Args) > 0 {
			argv = append(argv, os.Args[0])
		}

		argv = append(argv, i...)
	} else {
		argv = os.Args
	}

	if err := cmd.Run(argv); err != nil {
		panic(err)
	}
}
