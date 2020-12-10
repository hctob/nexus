package main

import (
  "log"
  "net/http"
  "github.com/gorilla/mux"
  "github.com/icza/session"
  "io/ioutil"
  "encoding/json"
)

type User struct{
  username string
  password string
}

func register(w http.ResponseWriter, r *http.Request){
  //put neo4j query here to add user
}

func login(w http.ResponseWriter, r *http.Request){
  //check login
  w.Header().Set("Content-Type", "application/json")
  body, _ := ioutil.ReadAll(r.Body)
  var user User
  err := json.Unmarshal(body, &user)
  if err != nil{
    log.Fatal(err)
  }
  //if password at node with username user.username equals user.password
    //log in
  result, err := session.Run(fmt.Sprintf("MATCH (n:Person {username: $u_name}) SET %s = $value RETURN %s", prop, prop), map[string]interface{}{
              "value": v,
              "u_name": un,
  })
  //get profile info from neo4j

  //create session
  sess := session.NewSession()
  session.Add(sess, w)
  sess := session.NewSessionOptions(&session.SessOptions{
    //username: user's username,
  })
}

func main(){
  http.Handle("/", http.FileServer(http.Dir("./nexus-frontend")))

  r := mux.NewRouter()
  r.HandleFunc("/register", register).Methods("POST")
  r.HandleFunc("/login", register).Methods("POST")

  http.ListenAndServe(":8080", nil)
}
