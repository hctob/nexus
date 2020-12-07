package main

import (
    "fmt"
    "flag"
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


type ChannelPool struct {
    createChannel  chan User     //channel for creating a new user
    created     chan bool
}

func pool_init() *ChannelPool {
    var cm ChannelPool
    cm.createChannel = make(chan User, 128)
    cm.created = make(chan bool)
    return &cm  //return pointer to newly initialized ChannelPool struct
}

func (cm ChannelPool) Listen() {
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

    driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth(username, password, ""), configForNeo4j40)
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
            fmt.Println("Create user message received")
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
            	fmt.Printf("Created Person '%s %s' with username = '%s'\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
            }
            cm.created <- true
        }
    }
}

/*
Runs a CREATE (n:Person) node with specified parameters
*/
/*func create_person(first_name, last_name, username, password string) error {
    result, err := session.Run("CREATE (n:Person { first_name: $first_name, last_name: $last_name, username: $username, password: $password}) RETURN n.first_name, n.last_name, n.username, n.password", map[string]interface{}{
    "first_name":   first_name,
    "last_name": last_name,
    "username": username,
    "password": password, })
    if err != nil {
        return err
    }
    for result.Next() {
    	fmt.Printf("Created Person '%s %s' with username = '%s'\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
    }
    return result.Err()
}*/

func (cm ChannelPool) create_person(first_name, last_name, username, password string) {
    var user = &User {first_name, last_name, username, password}
    cm.createChannel <- *user
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

    var cm ChannelPool
    cm.createChannel = make(chan User, 128)
    cm.created = make(chan bool)
    go drive("bolt://localhost:7687", *arg_username_raw, *arg_password_raw, cm)

    for {
        fmt.Println("Enter a create Person query: (first name, last name, username, password)")
        var first string
        fmt.Println("First name: ")
        fmt.Scanln(&first)

        var last string
        fmt.Println("Last name: ")
        fmt.Scanln(&last)

        var user string
        fmt.Println("Username: ")
        fmt.Scanln(&user)

        var pass string
        fmt.Println("Password: ")
        fmt.Scanln(&pass)

        cm.create_person(first, last, user, pass)
        created := <-cm.created
        if created == true {
            return
        }
    }
    fmt.Println("Done.")
}
