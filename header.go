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
    at_risk		string
	  last_infected_time	string
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

type House struct {
	username string
	address  string
	should_quarantine	string
}

type HouseRaw struct {
	address  string
	should_quarantine	string
}

type Join struct {
	target_user  string //username of Person to join the House of
	current_user string //username of Person currently logged in; used to create relationship for current_user->House
}

type JoinAddr struct {
	target_user  string //username of Person to join the House of
	address string //address of house to join
}

type HouseQuery struct {
	input      string
	isUsername bool //determines whether we are using a username or address as input
}

type Test struct {
	result	string
	date	string
	username string
}

type ChannelPool struct {
	createChannel     chan User //channel for creating a new user
	updateChannel     chan Update
	getNodeChannel    chan string
	friendChannel     chan Friends
	loginChannel      chan Login
	loginGood         chan bool
	loggedIn          chan User
	createHouse       chan House
	joinHouse         chan Join
	get_friends_list  chan string
	get_house         chan HouseQuery
	send_friends_list chan map[string]User
	notify_chan		chan string
	createHouseRaw	chan HouseRaw
	joinHouseAddr 	chan JoinAddr
	update_test		chan Test
}

func pool_init() ChannelPool {
  var cm ChannelPool
  cm.createChannel = make(chan User, 1024)
  cm.loginGood = make(chan bool)
  cm.updateChannel = make(chan Update, 1024)
  cm.getNodeChannel = make(chan string, 1024)
  cm.friendChannel = make(chan Friends, 1024)
  cm.loginChannel = make(chan Login, 1024)
  cm.loggedIn = make(chan User, 1024)
  cm.get_friends_list = make(chan string, 1024)
  cm.send_friends_list = make(chan map[string]User, 1024)
  cm.createHouse = make(chan House, 1024)
  cm.joinHouse = make(chan Join, 1024)
  cm.get_house = make(chan HouseQuery, 1024)
  cm.notify_chan = make(chan string, 1024)
  cm.createHouseRaw = make(chan HouseRaw, 1024)
  cm.joinHouseAddr = make(chan JoinAddr, 1024)
  cm.update_test = make(chan Test, 1024)
  return cm //return pointer to newly initialized ChannelPool struct
}
/*
*
* Creates a Person node with the specified properties
*/
func (cm ChannelPool) create_person(first_name, last_name, username, password string) {
    var user = &User {first_name, last_name, username, password, "false", "N/A"}
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

func (cm ChannelPool) create_house(username, address string) {
	house := &House{username, address, "false"}
	cm.createHouse <- *house
}

func (cm ChannelPool) create_house_raw(address string) {
	house := &HouseRaw{address, "false"}
	cm.createHouseRaw <- *house
}

func (cm ChannelPool) join_house(username, current_user string) {
	join := &Join{username, current_user}
	cm.joinHouse <- *join
}

func (cm ChannelPool) join_house_address(username, addr string) {
	join := &JoinAddr{username, addr}
	cm.joinHouseAddr <- *join
}

func (cm ChannelPool) get_friends(username string) {
	cm.get_friends_list <- username
	//usermap := make(map[string]User)
}

func (cm ChannelPool) get_household(input string, isUsername bool) {
	houseQuery := &HouseQuery{input, isUsername}
	cm.get_house <- *houseQuery
}

func (cm ChannelPool) notify_house(username string) {
	cm.notify_chan <- username
}

func (cm ChannelPool) add_test(result, date, username string) {
	test := &Test{result, date, username}
	cm.update_test <- *test
}

func print_user_info(user User) {
	fmt.Printf("\nUser: \"%s\"\nfirst_name: %s\nlast_name: %s\nat_risk: %s\nlast_infected_time: %s\n", user.username, user.first_name, user.last_name, user.at_risk, user.last_infected_time)
}

//connects a person to a house
/*func create_contact_house(u_name string, address string, s *neo4j.Session){
  result, _ := (*s).Run(`
        MATCH (a:Person),(h:House)
WHERE a.username = '$u_name' AND h.address = '$address'
CREATE (a)-[r:House]->(b) RETURN r`, nil)
}*/
