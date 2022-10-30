package main

import (
  "fmt"
  "log"
  "net/url"
  "net/http"
)

func main() {
  http.HandleFunc("/login", loginHandler)
  http.HandleFunc("/signup", signupHandler)
  http.HandleFunc("/shoppinglist", shoppinglistHandler)
  http.HandleFunc("/blog", blogHandler)
  http.HandleFunc("")
  log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
