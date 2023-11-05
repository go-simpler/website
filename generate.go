package main

import (
	"bufio"
	"fmt"
	"os"
	"text/template"
)

var tmpl = template.Must(template.New("").Parse(`<!doctype html>
<html>
    <head>
        <meta charset="utf-8" />
        <meta
            name="go-import"
            content="go-simpler.org/{{.}} git https://github.com/go-simpler/{{.}}"
        />
        <meta
            http-equiv="refresh"
            content="0; url=https://pkg.go.dev/go-simpler.org/{{.}}"
        />
    </head>
</html>
`))

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	pages, err := os.Open("pages.txt")
	if err != nil {
		return err
	}
	defer pages.Close()

	sc := bufio.NewScanner(pages)
	for sc.Scan() {
		name := sc.Text()
		f, err := os.Create(name + ".html")
		if err != nil {
			return err
		}
		if err := tmpl.Execute(f, name); err != nil {
			return err
		}
		_ = f.Close()
	}
	if err := sc.Err(); err != nil {
		return err
	}

	return nil
}
