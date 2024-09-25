package controllers

import "net/http"

type template interface {
	Execute(http.ResponseWriter, *http.Request, interface{})
}
