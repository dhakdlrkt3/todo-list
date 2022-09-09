package main

import (
	"log"
	"net/http"
	"yoojae-http/todos/app"
)

func main() {
	port := "3000"
	m := app.MakeHandler("./test.db")
	defer m.Close()

	log.Println("Started App")
	err := http.ListenAndServe(":"+port, m)
	if err != nil {
		panic(err)
	}
}
