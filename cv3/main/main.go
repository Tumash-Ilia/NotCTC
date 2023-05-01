package main

import (
	"fmt"
	"net/http"
	"os"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := "Ilia"
	hostname, _ := os.Hostname()
	fmt.Fprintf(w, "Hello, %s! Hostname: %s\n", name, hostname)
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	http.ListenAndServe(":8080", nil)
}
