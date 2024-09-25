package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

type Template struct {
	htmltpl *template.Template
}

func ParseFS(fs fs.FS, filePath ...string) (Template, error) {
	tpl := template.New(filePath[0])
	tpl.Funcs(
		template.FuncMap{
			"CSRFField": func() template.HTML {
				return `<input type=hidden>`
			},
		},
	)
	tpl, err := tpl.ParseFS(fs, filePath...)
	if err != nil {
		return Template{}, fmt.Errorf("ParseFS: %v", err)
	}
	return Template{
		htmltpl: tpl,
	}, nil
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}) {
	tpl, err := t.htmltpl.Clone()
	if err != nil {
		panic(err)
	}
	tpl.Funcs(
		template.FuncMap{
			"CSRFField": func() template.HTML {
				return csrf.TemplateField(r)
			},
		},
	)
	err = tpl.Execute(w, data)
	if err != nil {
		log.Printf("Execute: %v", err)
		http.Error(w, "Executing Problem", http.StatusInternalServerError)
	}
}

func Must(tpl Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return tpl
}
