package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
)

type Template struct {
	htmltpl *template.Template
}

func Parse(filePath ...string) (Template, error) {
	tpl, err := template.ParseFiles(filePath...)
	if err != nil {
		return Template{}, fmt.Errorf("Parse: %v", err)
	}
	return Template{
		htmltpl: tpl,
	}, err
}

func ParseFS(fs fs.FS, filePath ...string) (Template, error) {
	tpl, err := template.ParseFS(fs, filePath...)
	if err != nil {
		return Template{}, fmt.Errorf("ParseFS: %v", err)
	}
	return Template{
		htmltpl: tpl,
	}, nil
}

func (tpl Template) Execute(w http.ResponseWriter, data interface{}) {
	err := tpl.htmltpl.Execute(w, data)
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
