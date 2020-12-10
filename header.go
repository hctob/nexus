package main

import (
    "fmt"
    "github.com/neo4j/neo4j-go-driver/neo4j"
)
/*
* Any structs/helper functions
*/

type User struct {
    first_name string
    last_name string
    username string
    password string
}

type Update struct {
    username string
    property string
    value string
}

type Friends struct {
    u_name1 string
    u_name2 string
}

type Login struct {
    username string
    password string
}


type ChannelPool struct {
    createChannel  chan User     //channel for creating a new user
    updateChannel chan Update
    getNodeChannel   chan string
    friendChannel   chan Friends
    loginChannel    chan Login
    loginGood     chan bool
}

func pool_init() ChannelPool {
    var cm ChannelPool
    cm.createChannel = make(chan User, 128)
    cm.loginGood = make(chan bool)
    cm.updateChannel = make(chan Update, 128)
    cm.getNodeChannel = make(chan string, 128)
    cm.friendChannel = make(chan Friends, 128)
    cm.loginChannel = make(chan Login, 128)
    return cm  //return pointer to newly initialized ChannelPool struct
}
/*
*
* Creates a Person node with the specified properties
*/
func (cm ChannelPool) create_person(first_name, last_name, username, password string) {
    var user = &User {first_name, last_name, username, password}
    cm.createChannel <- *user
}

func (cm ChannelPool) get_person(username string) {
    cm.getNodeChannel <- username
}

func (cm ChannelPool) make_friends(u_name1, u_name2 string) {
    var friend = &Friends { u_name1, u_name2 }
    cm.friendChannel <- *friend
}

func create_contact_person(u_name1 string, u_name2 string, s *neo4j.Session){
  _, err := (*s).Run("MATCH (a:Person),(b:Person) WHERE a.username = $u_name1 AND b.username = $u_name2 CREATE (a)-[r:Friend]->(b) RETURN r", map[string]interface{}{
      "u_name1" : u_name1,
       "u_name2" : u_name2,
  })

if err != nil {
    fmt.Println("Error:\n", err)
    return
}
fmt.Println("friend relationship created\n")

}

func (cm ChannelPool) update_by_username(u_name, property, value string) {
    var up = &Update {u_name, property, value}
    cm.updateChannel <- *up
}

//update a user's property
func update_user(u_name string, property string, value string, s *neo4j.Session) {
  result, err := (*s).Run("MATCH (a:Person{username: $u_name}) SET a.$property = $value RETURN a", map[string]interface{}{
    "u_name": u_name,
    "property":   property,
    "value": value, })

if err != nil {
    fmt.Println("Error:\n", err)
    return
}
//fmt.Println("query fine\n")
fmt.Printf("Updated %s %s to %s\n\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
}

func (cm ChannelPool) login(username, password string) bool {
    logInfo := &Login{username, password}
    cm.loginChannel <- *logInfo
    good := <-cm.loginGood
    return good
}

//connects a person to a house
/*func create_contact_house(u_name string, address string, s *neo4j.Session){
  result, _ := (*s).Run(`
        MATCH (a:Person),(h:House)
WHERE a.username = '$u_name' AND h.address = '$address'
CREATE (a)-[r:House]->(b) RETURN r`, nil)
}*/
