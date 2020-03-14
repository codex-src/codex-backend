package main

import (
	"net/http"
	"strings"
)

const PUBLIC_URL = "http://localhost:3000"

var Methods = []string{
	"GET",
	"OPTIONS",
	"POST",
}

var Headers = []string{
	"Access-Control-Allow-Headers", // Needed?
	"Authorization",
	"Content-Type",
	"X-Requested-With",
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(Headers, ", "))
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(Methods, ", "))
	w.Header().Set("Access-Control-Allow-Origin", PUBLIC_URL)
}
