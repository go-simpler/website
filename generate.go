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
                {{- range $.Projects}}
                {{- if eq .Category $category}}
                <tr>
                    <td><a href="{{.Href}}">{{.Name}}</a></td>
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
	data, err := os.ReadFile("packages.txt")
	if err != nil {
		return err
	}

	if err := os.Chdir("docs"); err != nil {
		return err
	}

	packages := strings.Split(strings.TrimSpace(string(data)), "\n")
	misc := [...]string{"styleguide", "website", ".github"}
	categories := [...]string{"libs", "tools", "misc"}
	projects := make([]struct{ Name, Desc, Href, Category string }, len(packages)+len(misc))

	for i, name := range packages {
		desc, topics, err := getRepoInfo(name)
		if err != nil {
			return err
		}

		projects[i].Name = "go-simpler.org/" + name
		projects[i].Desc = desc
		projects[i].Href = name + ".html"

		switch {
		case slices.Contains(topics, "library"):
			projects[i].Category = categories[0]
		case slices.Contains(topics, "tool"):
			projects[i].Category = categories[1]
		default:
			return fmt.Errorf("%s: no category set", name)
		}

		f, err := os.Create(name + ".html")
		if err != nil {
			return err
		}
		defer f.Close()

		if err := pageTmpl.Execute(f, name); err != nil {
			return err
		}
	}

	for i, name := range misc {
		desc, _, err := getRepoInfo(name)
		if err != nil {
			return err
		}

		j := len(packages) + i
		projects[j].Name = "github.com/go-simpler/" + name
		projects[j].Desc = desc
		projects[j].Href = "https://github.com/go-simpler/" + name
		projects[j].Category = categories[2]
	}

	f, err := os.Create("index.html")
	if err != nil {
		return err
	}
	defer f.Close()

	return indexTmpl.Execute(f, struct {
		Projects   any
		Categories []string
	}{
		Projects:   projects,
		Categories: categories[:],
	})
}

func getRepoInfo(name string) (desc string, topics []string, _ error) {
	resp, err := http.Get("https://api.github.com/repos/go-simpler/" + name)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	var v struct {
		Desc   string   `json:"description"`
		Topics []string `json:"topics"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", nil, err
	}

	return v.Desc, v.Topics, nil
}
