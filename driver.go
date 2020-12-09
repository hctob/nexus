package main

import (
    "fmt"
    "flag"
    "time"
    "runtime"
    "github.com/neo4j/neo4j-go-driver/neo4j"
)

/*
*
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


type ChannelPool struct {
    createChannel  chan User     //channel for creating a new user
    updateChannel chan Update
    getNodeChannel   chan string
    created     chan bool
}

func pool_init() *ChannelPool {
    var cm ChannelPool
    cm.createChannel = make(chan User, 128)
    cm.created = make(chan bool)
    return &cm  //return pointer to newly initialized ChannelPool struct
}

var (
	arg_uri     = flag.String("uri", "bolt://localhost:7687", "The URI for the Nexus database, to connect to it.")
    arg_username_raw     = flag.String("u", "test", "Usernames are unique identifiers for database users.")
    arg_password_raw     = flag.String("p", "test", "Unencrypted password for selected username.")

    totalQueries int64
)


func drive(uri, username, password string, cm ChannelPool) {
        // configForNeo4j35 := func(conf *neo4j.Config) {}
    configForNeo4j40 := func(conf *neo4j.Config) { conf.Encrypted = false }

    driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""), configForNeo4j40)
    if err != nil {
		fmt.Println("Error:", err)
        return
	}
    //fmt.Println("established driver connection\n")
    // handle driver lifetime based on your application lifetime requirements
    // driver's lifetime is usually bound by the application lifetime, which usually implies one driver instance per application
    defer driver.Close()

    // For multidatabase support, set sessionConfig.DatabaseName to requested database
    sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
    session, err := driver.NewSession(sessionConfig)
    if err != nil {
		fmt.Println("Error:\n", err)
        return
	}
    //fmt.Println("established session connection\n")
    defer session.Close()

    for {
        select {
        case user := <-cm.createChannel:
            //fmt.Println("Create user message received")
            result, err := session.Run("CREATE (n:Person { first_name: $first_name, last_name: $last_name, username: $username, password: $password}) RETURN n.first_name, n.last_name, n.username, n.password", map[string]interface{}{
            "first_name":   user.first_name,
            "last_name": user.last_name,
            "username": user.username,
            "password": user.password, })

            if err != nil {
                fmt.Println("Error:\n", err)
                return
            }
            //fmt.Println("query fine\n")
            for result.Next() {
            	fmt.Printf("Created Person '%s %s' with username = '%s'\n\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
            }
            //cm.created <- true
        case update := <-cm.updateChannel:
            //update_user(update.username, update.property, update.value, &session)
            prop := "n." + update.property
            fmt.Println(prop)
            result, err := session.Run("match (n:Person {username: $u_name}) set $property = $value, return $property", map[string]interface{}{
              "u_name": update.username,
              "property":   prop,
              "value": update.value, })
            fmt.Printf("match (n:Person {username: %s}) set %s = %s return %s", update.username, prop, update.value, prop)

              if err != nil {
                  fmt.Println("Error:\n", err)
                  return
              }
              //fmt.Println("query fine\n")
              fmt.Printf("Updated %s %s to %s\n\n", update.username, update.property, result.Record().GetByIndex(0).(string))
        case username := <-cm.getNodeChannel:
            result, err := session.Run("match (n:Person {username: $u_name}) return n", map[string]interface{}{
              "u_name": username, })

              if err != nil {
                  fmt.Println("Error:\n", err)
                  return
              }
              //fmt.Println("query fine\n")
              for result.Next() {
                  fmt.Printf("Node: %s\n\n", result.Record().GetByIndex(0))
              }
        }
    }
}

/*
* Creates a Person node with the specified properties
*/
func (cm ChannelPool) create_person(first_name, last_name, username, password string) {
    var user = &User {first_name, last_name, username, password}
    cm.createChannel <- *user
}

func (cm ChannelPool) get_person(username string) {
    cm.getNodeChannel <- username
}

func create_contact_person(u_name1 string, u_name2 string, s *neo4j.Session){
  result, err := (*s).Run(`
        MATCH (a:Person),(b:Person)
WHERE a.username = '$u_name1' AND b.username = '$u_name2'
CREATE (a)-[r:Friend]->(b) RETURN r`, nil)

if err != nil {
    fmt.Println("Error:\n", err)
    return
}
//fmt.Println("query fine\n")
for result.Next() {
    fmt.Printf("Updated Person %s %s' with username = '%s'\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
}

}

//connects a person to a house
/*func create_contact_house(u_name string, address string, s *neo4j.Session){
  result, _ := (*s).Run(`
        MATCH (a:Person),(h:House)
WHERE a.username = '$u_name' AND h.address = '$address'
CREATE (a)-[r:House]->(b) RETURN r`, nil)
}*/

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

/*
* Main function for driver
*/
func main() {
    runtime.GOMAXPROCS(256)
	flag.Parse()

    fmt.Println("Nexus Command Line Driver: ")
    if *arg_username_raw == "" {
        fmt.Println("god damn it\n")
        return
    }
    if *arg_password_raw == "" {
        fmt.Println("god damn it\n")
        return
    }

    //TODO: use pool_init()
    var cm ChannelPool
    cm.createChannel = make(chan User, 128)
    cm.created = make(chan bool)
    cm.updateChannel = make(chan Update, 128)
    cm.getNodeChannel = make(chan string, 128)

    go drive("bolt://localhost:7687", *arg_username_raw, *arg_password_raw, cm)

    for {

        var option string
        fmt.Println("\nOptions: ")
        fmt.Println("1. Create a user (1, create): \n2. Update a specific property of a Person node (update)\n3. Get a node by username\n4. Exit")
        fmt.Printf("nexus> ")
        fmt.Scanln(&option)
        if option == "create"  || option == "1" {
            fmt.Println("Enter a create Person query: (first name, last name, username, password)")
            var first string
            fmt.Println("First name: ")
            fmt.Printf("nexus> ")
            fmt.Scanln(&first)

            var last string
            fmt.Println("Last name: ")
            fmt.Printf("nexus> ")
            fmt.Scanln(&last)

            var user string
            fmt.Println("Username: ")
            fmt.Printf("nexus> ")
            fmt.Scanln(&user)

            var pass string
            fmt.Println("Password: ")
            fmt.Printf("nexus> ")
            fmt.Scanln(&pass)

            cm.create_person(first, last, user, pass)
            time.Sleep(time.Millisecond * 500)

        } else if option == "update" || option == "2"{
            var user string
            fmt.Println("Username: ")
            fmt.Printf("nexus> ")
            fmt.Scanln(&user)

            var property string
            fmt.Println("Property: ")
            fmt.Printf("nexus> ")
            fmt.Scanln(&property)

            var value string
            fmt.Println("Value: ")
            fmt.Printf("nexus> ")
            fmt.Scanln(&value)
            //var update = &Update {user, property, value}
            cm.update_by_username(user, property, value)
            time.Sleep(time.Millisecond * 500)

        } else if option == "e" || option == "exit" || option == "4" {
            fmt.Println("Exiting... gracefully")
            return
        } else if option == "get" || option == "3" {
            var user string
            fmt.Println("Username: ")
            fmt.Printf("nexus> ")
            fmt.Scanln(&user)
            cm.get_person(user)
            time.Sleep(time.Millisecond * 500)
        }
    }
    fmt.Println("Done.")
}
