package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
)

var indexTmpl = template.Must(template.New("").Parse(`<!doctype html>
<html>
    <head>
        <meta charset="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <link rel="stylesheet" href="styles.css" />
        <title>go-simpler.org</title>
    </head>

    <body>
        <div id="content">
            <div class="center">
                <h2>go-simpler.org</h2>
                <p>A collection of Go packages built with ❤️</p>
            </div>
            <table>
                {{- range .}}
                <tr>
                    <td><a href="{{.Name}}.html">go-simpler.org/{{.Name}}</a></td>
                    <td>{{.Desc}}</td>
                </tr>
                {{- end}}
            </table>
        </div>
    </body>
</html>
`))

var pageTmpl = template.Must(template.New("").Parse(`<!doctype html>
<html>
    <head>
        <meta charset="utf-8" />
        <meta name="go-import" content="go-simpler.org/{{.}} git https://github.com/go-simpler/{{.}}" />
        <meta http-equiv="refresh" content="0; url=https://pkg.go.dev/go-simpler.org/{{.}}" />
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
	data, err := os.ReadFile("pages.txt")
	if err != nil {
		return err
	}

	if err := os.Chdir("docs"); err != nil {
		return err
	}

	names := strings.Split(strings.TrimSpace(string(data)), "\n")
	pages := make([]struct{ Name, Desc string }, len(names))

	for i, name := range names {
		err := func() error {
			resp, err := http.Get("https://api.github.com/repos/go-simpler/" + name)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var v struct {
				Desc string `json:"description"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
				return err
			}

			pages[i].Name = name
			pages[i].Desc = v.Desc

			f, err := os.Create(name + ".html")
			if err != nil {
				return err
			}
			defer f.Close()

			return pageTmpl.Execute(f, name)
		}()
		if err != nil {
			return err
		}
	}

	f, err := os.Create("index.html")
	if err != nil {
		return err
	}
	defer f.Close()

	return indexTmpl.Execute(f, pages)
}
