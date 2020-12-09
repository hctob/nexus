package main

import (
  "net/http"
)

func main(){
  http.Handle("/", http.FileServer(http.Dir("./nexus-frontend")))
  http.ListenAndServe(":10543", nil)
}
