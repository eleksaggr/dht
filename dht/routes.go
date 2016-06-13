package main

import "net/http"

type Route struct {
	Name    string
	Method  string
	Path    string
	Handler http.HandlerFunc
}

var routes = []Route{
	Route{
		Name:    "Get",
		Method:  "GET",
		Path:    "/{key}",
		Handler: HandleGet,
	},
	Route{
		Name:    "Set",
		Method:  "POST",
		Path:    "/",
		Handler: HandleSet,
	},
	Route{
		Name:    "Delete",
		Method:  "DELETE",
		Path:    "/{key}",
		Handler: HandleDelete,
	},
}
