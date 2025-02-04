package main

import (
	"os"
	"strings"
	"text/template"
)

type Formula struct {
	Owner          string
	Name           string
	Version        string
	Tag            string
	DarwinArm64SHA string
	DarwinAmd64SHA string
	LinuxArm64SHA  string
	LinuxAmd64SHA  string
}

func main() {
	repo := os.Getenv("GITHUB_REPOSITORY")
	owner, name, _ := strings.Cut(repo, "/")
	tag := os.Getenv("VERSION")
	version := strings.TrimPrefix(tag, "v")

	formula := Formula{
		Owner:          owner,
		Name:           name,
		Version:        version,
		Tag:            tag,
		DarwinArm64SHA: os.Getenv("DARWIN_ARM64"),
		DarwinAmd64SHA: os.Getenv("DARWIN_AMD64"),
		LinuxArm64SHA:  os.Getenv("LINUX_ARM64"),
		LinuxAmd64SHA:  os.Getenv("LINUX_AMD64"),
	}

	tmpl := template.Must(template.ParseFiles("homebrew/formula.rb.tpl"))
	if err := tmpl.Execute(os.Stdout, formula); err != nil {
		panic(err)
	}
}
