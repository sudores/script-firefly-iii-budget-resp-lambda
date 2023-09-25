package main

import (
	"net/http"
	"os"
)

func main() {
	ffiToken := os.Getenv("FIREFLY_III_API_KEY")
	budgets := map[string]string{
		"food": "",
	}

	http.HandleFunc("/food", nil)
	http.HandleFunc("/cafe", nil)

	http.ListenAndServe(":3000", nil)
}
