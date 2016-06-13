package main

import (
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/zillolo/dht"
)

var server *dht.Node

type pair struct {
	Key   string
	Value string
}

func InitAPI(host string, node *dht.Node) {
	if node == nil {
		log.Printf("Error during initalization of API.\n")
		return
	}
	server = node

	// router := NewRouter()
	// http.ListenAndServe(host, router)
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/keys", HandleGetAll),
		rest.Get("/keys/:key", HandleGet),
		rest.Post("/keys", HandleSet),
		rest.Delete("/keys/:key", HandleDelete),
	)
	if err != nil {
		log.Fatalf("%v\n", err)
		return
	}

	api.SetApp(router)

	http.Handle("/", api.MakeHandler())
	http.HandleFunc("/status", HandleStatus)

	http.ListenAndServe(host, nil)
}

func HandleSet(w rest.ResponseWriter, r *rest.Request) {
	var p pair
	if err := r.DecodeJsonPayload(&p); err != nil {
		log.Fatalf("Error during decoding of JSON payload.\n")
		rest.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if p.Key == "" {
		rest.Error(w, "country code required", 400)
		return
	}
	if p.Value == "" {
		rest.Error(w, "country name required", 400)
		return
	}
	if err := server.Role().(*dht.Leader).Set(p.Key, p.Value); err != nil {
		error := map[string]interface{}{"error": err.Error()}
		if err := w.WriteJson(&error); err != nil {
			log.Fatalf("%v\n", err)
		}
	} else {
		response := map[string]interface{}{"success": true}
		if err := w.WriteJson(&response); err != nil {
			log.Fatalf("%v\n", err)
		}
	}

}

func HandleGet(w rest.ResponseWriter, r *rest.Request) {
	key := r.PathParam("key")

	value, err := server.Role().(*dht.Leader).Get(key)
	if err != nil {
		rest.NotFound(w, r)
		return
	} else {
		response := map[string]interface{}{"value": value, "success": true}
		if err := w.WriteJson(&response); err != nil {
			log.Fatalf("%v\n", err)
		}
	}
}

func HandleGetAll(w rest.ResponseWriter, r *rest.Request) {
	rest.Error(w, "", 501)
}

func HandleStatus(w http.ResponseWriter, r *http.Request) {
}

func HandleDelete(w rest.ResponseWriter, r *rest.Request) {
	key := r.PathParam("key")

	err := server.Role().(*dht.Leader).Delete(key)
	if err != nil {
		rest.NotFound(w, r)
		return
	} else {
		response := map[string]interface{}{"success": true}
		if err := w.WriteJson(&response); err != nil {
			log.Fatalf("%v\n", err)
		}
	}
}
