package views

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"third/context"
	"third/models"

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
			"currentUser": func() (*models.User, error) {
				return nil, fmt.Errorf("current user not implemented")
			},
			"errors": func() []string {
				return nil
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

func Must(tpl Template, err error) Template {
	if err != nil {
		panic(err)
	}
	return tpl
}

type public interface {
	Public() string
}

func errorMessages(errs ...error) []string {
	var msgs []string
	for _, err := range errs {
		var pubErr public
		if errors.As(err, &pubErr) {
			msgs = append(msgs, pubErr.Public())
		} else {
			msgs = append(msgs, "Something went wrong")
		}
	}
	return msgs
}

func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}, errs ...error) {
	tpl, err := t.htmltpl.Clone()
	if err != nil {
		panic(err)
	}
	tpl.Funcs(
		template.FuncMap{
			"CSRFField": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() (*models.User, error) {
				return context.User(r.Context()), nil
			},
			"errors": func() []string {
				return errorMessages(errs...)
			},
		},
	)
	err = tpl.Execute(w, data)
	if err != nil {
		log.Printf("Execute: %v", err)
		http.Error(w, "Executing Problem", http.StatusInternalServerError)
	}
}
