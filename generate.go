package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"slices"
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
                {{- range $_, $category := $.Categories}}
                <tr>
                    <th colspan="2">{{.}}</th>
                </tr>
                {{- range $.Packages}}
                {{- if eq .Category $category}}
                <tr>
                    <td><a href="{{.Name}}.html">go-simpler.org/{{.Name}}</a></td>
                    <td>{{.Desc}}</td>
                </tr>
                {{- end}}
                {{- end}}
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
	packages := make([]struct{ Name, Desc, Category string }, len(names))
	categories := [...]string{"libs", "tools"}

	for i, name := range names {
		err := func() error {
			resp, err := http.Get("https://api.github.com/repos/go-simpler/" + name)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			var v struct {
				Desc   string   `json:"description"`
				Topics []string `json:"topics"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
				return err
			}

			packages[i].Name = name
			packages[i].Desc = v.Desc

			switch {
			case slices.Contains(v.Topics, "library"):
				packages[i].Category = categories[0]
			case slices.Contains(v.Topics, "tool"):
				packages[i].Category = categories[1]
			default:
				return fmt.Errorf("%s: no category set", name)
			}

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

	return indexTmpl.Execute(f, struct {
		Packages   any
		Categories []string
	}{
		Packages:   packages,
		Categories: categories[:],
	})
}
