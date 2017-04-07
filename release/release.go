package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var tmpl = `class {{.Name | Title }} < Formula
  desc "{{.Description}}"
  homepage "{{.Homepage}}"
  url "{{.Zip}}"
  version "{{.Version}}"
  sha256 "{{.Sha256}}"

  depends_on :arch => :intel

  def install
    bin.install Dir['*']
  end

  test do
    system "false"
  end
end
`

func main() {

	if len(os.Args) != 3 {
		log.Println("Usage: go run main.go VERSION FILE")
		os.Exit(0)
	}

	name := "rtail"
	version := os.Args[1]
	file := os.Args[2]

	file, err := filepath.Abs(file)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	f, err := os.Open(file)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	checkSum := sha256.Sum256(buf)

	tmpl, err := template.New("formula").Funcs(template.FuncMap{
		"Title": strings.Title,
	}).Parse(tmpl)
	if err != nil {
		log.Fatal(err)
	}

	if err := tmpl.Execute(os.Stdout, struct {
		Name, Version, Sha256, Description, Homepage, Zip string
	}{
		Name:        name,
		Version:     version,
		Sha256:      fmt.Sprintf("%x", checkSum),
		Description: "Remote tailing tool",
		Homepage:    "https://github.com/chonthu/" + name,
		Zip:         fmt.Sprintf("https://github.com/chonthu/%s/releases/download/%s/%s.gz", name, version, name),
	}); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
