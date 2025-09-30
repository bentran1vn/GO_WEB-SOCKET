package main

import (
	"fmt"
	"net/http"
)

func main() {
	setupAPI()

	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("server error:", err)
	}
}

func setupAPI() {
	manager := NewManager()

	http.Handle("/", http.FileServer(http.Dir("../frontend")))
	http.HandleFunc("/ws", manager.ServeWS)
}
