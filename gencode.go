//go:build ignore

package main

import (
    "fmt"
    "strings"
    "io/ioutil"
    "os"
    "text/template"
    "path"

    "github.com/go-yaml/yaml"
)

func main() {
    funcmap := template.FuncMap{
        "title": strings.Title,
    }
    data, err := ioutil.ReadFile(os.Args[2])
    if err != nil {
        panic(err)
    }
    var schemas []map[string]any
    if err = yaml.Unmarshal(data, &schemas); err != nil {
        panic(err)
    }
    for _, a := range schemas {
        t := template.New(path.Base(os.Args[1])).Funcs(funcmap)
        t = template.Must(t.ParseFiles(os.Args[1]))
        os.Mkdir(path.Join(os.Args[3], a["name"].(string)), os.ModePerm)
        f, err := os.Create(path.Join(os.Args[3], a["name"].(string), fmt.Sprintf("%s_gen.go", a["name"])))
        if err != nil {
            panic(err)
        }
        if err = t.Execute(f, a); err != nil {
            panic(err)
        }
        if err = f.Close(); err != nil {
            panic(err)
        }
    }
}
