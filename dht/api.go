package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zillolo/dht"
)

var server *dht.Node

func InitAPI(host string, node *dht.Node) {
	if node == nil {
		log.Printf("Error during initalization of API.\n")
		return
	}
	server = node

	router := NewRouter()
	http.ListenAndServe(host, router)
}

func HandleSet(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}

	if err = r.Body.Close(); err != nil {
		panic(err)
	}

	var data map[string]interface{}
	if err = json.Unmarshal(body, &data); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err = json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}

	key := data["key"].(string)
	value := data["value"].(string)

	err = server.Role().(*dht.Leader).Set(key, value)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte("Success\n"))
	}
}

func HandleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	key := vars["key"]

	key, err := server.Role().(*dht.Leader).Get(key)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(key))
	}
}

func HandleDelete(w http.ResponseWriter, r *http.Request) {
	// body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	// if err != nil {
	// 	panic(err)
	// }
	//
	// if err = r.Body.Close(); err != nil {
	// 	panic(err)
	// }
	//
	// var data map[string]interface{}
	// if err = json.Unmarshal(body, &data); err != nil {
	// 	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// 	w.WriteHeader(422) // unprocessable entity
	// 	if err = json.NewEncoder(w).Encode(err); err != nil {
	// 		panic(err)
	// 	}
	// }
	//
	// key := data["key"].(string)
	// value := data["value"].(string)
	// message := dht.Message{
	// 	Action: dht.Message_DELETE.Enum(),
	// 	Key:    &key,
	// 	Value:  &value,
	// }
	//
	// err = server.Role().(*dht.Leader).Delete(key)
	// if err != nil {
	// 	w.Write([]byte(err.Error()))
	// } else {
	// 	w.Write([]byte("Success\n"))
	// }
}
