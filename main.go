package main

import "github.com/brawaru/marct/cmd"

//go:generate goi18n extract -sourceLanguage en-US -outdir locales

func main() {
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
