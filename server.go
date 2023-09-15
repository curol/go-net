package main

import (
	"request"
	"server"
	"store"
	"writer"
)

var db = store.NewStore()

func main() {
	// Create server
	server := server.NewServer("tcp", ":8080", nil)

	// Set Handlers
	server.GET("/ping", pingHandler)
	server.POST("/", postHandler)
	server.GET("/", getHandler)
	server.NotFound("/", notFoundHandler)

	// Start server
	server.Run()
}

// Handlers
func notFoundHandler(req *request.Request, w *writer.ResponseWriter) {
	w.Write([]byte("404 Not Found"))
}

func pingHandler(r *request.Request, w *writer.ResponseWriter) {
	w.Write([]byte("PONG"))
}

func getHandler(req *request.Request, w *writer.ResponseWriter) {
	// validate

	// read and decode request's body
	// req.JSON(result)

	// get from db
	// result := db.Get(req.Path())

	// Response
	w.Write([]byte("GET"))
}

func postHandler(req *request.Request, w *writer.ResponseWriter) {
	// var kv KV
	// err := req.JSON(&kv)
	// fmt.Println("KV: ", kv)
	// if err != nil {
	// 	w.Write([]byte(err.Error()))
	// 	return
	// }

	// Db.Set(kv.Key, kv.Value)
	// v := Db.Get(kv.Key)

	// w.Write([]byte(v))

	w.Write([]byte("POST"))
}
