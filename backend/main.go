package main

import (
	"context"
	"fmt"
	"net/http"
)

func main() {
	setupAPI()

	fmt.Println("Listening on :8080")
	if err := http.ListenAndServeTLS(":8080", "server.crt", "server.key", nil); err != nil {
		fmt.Println("server error:", err)
	}
}

func setupAPI() {

	ctx := context.Background()

	manager := NewManager(ctx)

	http.Handle("/", http.FileServer(http.Dir("../frontend")))
	http.HandleFunc("/ws", manager.ServeWS)
	http.HandleFunc("/login", manager.loginHandler)
}
