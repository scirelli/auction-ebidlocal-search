package main

import (
	"html/template"
	"log"
	"os"

	"github.com/scirelli/auction-ebidlocal-search/internal/pkg/ebidlocal"
)

func main1() {
	t, err := template.New("template.html.tmpl").Funcs(template.FuncMap{
		"htmlSafe": func(html string) template.HTML {
			return template.HTML(html)
		},
	}).ParseFiles("./assets/templates/template.html.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	var keywords = ebidlocal.Keywords{"nintendo", "sega", "chainsaw", "turbografx", "playstation", "ps4", "ps3", "famicom"}

	t.Execute(os.Stdout, keywords.Search())
}
