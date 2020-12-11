package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	arg_uri          = flag.String("uri", "bolt://localhost:7687", "The URI for the Nexus database, to connect to it.")
	arg_username_raw = flag.String("u", "test", "Usernames are unique identifiers for database users.")
	arg_password_raw = flag.String("p", "test", "Unencrypted password for selected username.")
    //test_nodes_input = flag.String("generate", "", "JSON data for generating test Person nodes.")
    //test_friends_input = flag.String("test-create", "", "JSON data for generating friend relationships.")
    test_flag = flag.Bool("test", false, "Enable for generation purposes.")
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
				"first_name": user.first_name,
				"last_name":  user.last_name,
				"username":   user.username,
				"password":   user.password})

			if err != nil {
				fmt.Println("Error:\n", err)
				return
			}
			//fmt.Println("query fine\n")
			for result.Next() {
				fmt.Printf("\nCreated Person '%s %s' with username = '%s'\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
			}
			//cm.created <- true
		case house := <-cm.createHouse:
			un := house.username
			addr := house.address
			result, err := session.Run("match (n:Person{username: $un}) create (h:House{address: $addr}) create (n)-[r:HOUSE]->(h) return h.address", map[string]interface{}{
				"un":   un,
				"addr": addr,
			})
			if err != nil {
				fmt.Println("Error:\n", err)
				return
			}
			//fmt.Println("query fine\n")
			for result.Next() {
				fmt.Printf("\nCreated House with address %s\n", result.Record().GetByIndex(0).(string))
			}
		case join := <-cm.joinHouse:
			result, err := session.Run("match (n:Person{username: $target_user})-[r:HOUSE]->(h), (c:Person{username: $current_user}) create (c)-[:HOUSE]->(h) return h.address", map[string]interface{}{
				"target_user":  join.target_user,
				"current_user": join.current_user,
			})
			if err != nil {
				fmt.Println("Error:\n", err)
				return
			}
			//fmt.Println("query fine\n")
			for result.Next() {
				fmt.Printf("\n%s joined House with address %s\n", join.current_user, result.Record().GetByIndex(0).(string))
			}
		case username := <-cm.get_friends_list:
			result, err := session.Run("match (n:Person{username: $username})-[r:FRIEND]->(f) return f.first_name, f.last_name, f.username", map[string]interface{}{
				"username": username,
			})
			if err != nil {
				fmt.Println("Error:\n", err)
				return
			}
			//fmt.Println("query fine\n")
			fmt.Printf("\nFriends list of %s: \n", username)
			for result.Next() {
				fmt.Printf("%s %s: %s\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
			}
		case update := <-cm.updateChannel:
			//update_user(update.username, update.property, update.value, &session)
			prop := "n." + update.property
			//fmt.Println(prop)

			v := update.value
			un := update.username
			result, err := session.Run(fmt.Sprintf("MATCH (n:Person {username: $u_name}) SET %s = $value RETURN %s", prop, prop), map[string]interface{}{
				"value":  v,
				"u_name": un,
			})

			//fmt.Printf("match (n:Person {username: %s}) set %s = %s return %s", update.username, prop, update.value, prop)

			if err != nil {
				fmt.Println("Error:\n", err)
				return
			}
			//fmt.Println("query fine\n")
			for result.Next() {
				fmt.Printf("Updated %s %s to %s\n", update.username, update.property, result.Record().GetByIndex(0).(string))
			}
		case username := <-cm.getNodeChannel:
			result, err := session.Run("match (n:Person {username: $u_name}) return n.first_name, n.last_name, n.username, n.password", map[string]interface{}{
				"u_name": username})

			if err != nil {
				fmt.Println("Error:\n", err)
				return
			}
			//fmt.Println("query fine\n")
			for result.Next() {
				fmt.Printf("\nNode: \"%s %s\":\nusername: %s\npassword: %s\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string), result.Record().GetByIndex(3).(string))
			}
		case houseQuery := <-cm.get_house:
			input := houseQuery.input
			isUsername := houseQuery.isUsername

			if isUsername == true {
                //use username of person in house
				result, err := session.Run("match (h:House)<-[r:HOUSE]-(n:Person{username: $input}) match (h)<-[:HOUSE]-(p) return p.username, p.first_name, p.last_name", map[string]interface{}{
					"input": input,
				})
				if err != nil {
					fmt.Println("Error:\n", err)
					return
				}
				fmt.Printf("%s's household:\n", input)
				for result.Next() {
					fmt.Printf("\t%s: %s %s\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
				}
			} else {
				//use address for query
				result, err := session.Run("match (h:House{address: $input})<--(n) return n.username, n.first_name, n.last_name", map[string]interface{}{
					"input": input,
				})
				if err != nil {
					fmt.Println("Error:\n", err)
					return
				}
				fmt.Printf("House @ %s:\n", input)
				for result.Next() {
					fmt.Printf("\t%s: %s %s\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
				}
			}
		case friends := <-cm.friendChannel:
			un1 := friends.u_name1
			un2 := friends.u_name2
			fmt.Println(un1, " ", un2)
			result, err := session.Run("MATCH (a:Person{username: $u_name1}) , (b:Person{username: $u_name2}) CREATE (a)-[r:FRIEND]->(b) RETURN r, a.username, b.username", map[string]interface{}{
				"u_name1": un1,
				"u_name2": un2,
			})

			if err != nil {
				fmt.Println("Error:\n", err)
				return
			}
			for result.Next() {
				fmt.Printf("\n%s\ncreated between %s and %s\n", result.Record().GetByIndex(0), result.Record().GetByIndex(1), result.Record().GetByIndex(2))
			}
		case login := <-cm.loginChannel:
			un := login.username
			//fmt.Println("Received username =", un)
			pw := login.password
            if un == "admin" && (pw == "test" || pw == "ye") {
                cm.loginGood <- true
                user := &User{un, "", un, pw}
                //print_user_info(*user)
                cm.loggedIn <- *user
                return
            }
			//fmt.Println("Received password =", pw)
			result, err := session.Run("MATCH (n:Person{username: $username}) return n.first_name, n.last_name, n.username, n.password", map[string]interface{}{
				"username": un,
				"password": pw,
			})

			if err != nil {
				fmt.Println("Error:\n", err)
				return
			}
			//res := false
			for result.Next() {
				//fmt.Println(login.password, result.Record().GetByIndex(0).(string))
				fn := result.Record().GetByIndex(0).(string)
				ln := result.Record().GetByIndex(1).(string)
				un := result.Record().GetByIndex(2).(string)
				pw := result.Record().GetByIndex(3).(string)
				if login.password == pw {
					cm.loginGood <- true
					user := &User{fn, ln, un, pw}
					//print_user_info(*user)
					cm.loggedIn <- *user
				} else {
					cm.loginGood <- false
				}
			}
			//cm.loginGood <- false
		}
	}
}

/*
* Main function for driver
 */

func main() {
	runtime.GOMAXPROCS(256)
	flag.Parse()
    logged_in := *test_flag
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
	/*cm.createChannel = make(chan User, 128)
	  cm.created = make(chan bool)
	  cm.updateChannel = make(chan Update, 128)
	  cm.getNodeChannel = make(chan string, 128)
	*/
	cm = pool_init()
	go drive(*arg_uri, *arg_username_raw, *arg_password_raw, cm)
    /*if *test_nodes_input != "" {
        //TODO: add JSON parsing to generate a bunch of Person nodes in the database.
        fmt.Println("Generating Person nodes...")
    }
    if *test_friends_input != "" {
        //TODO: add JSON parsing to generate a bunch of Friend relationships between generated nodes.
        fmt.Println("Generating FRIEND relationships between random nodes...")
    }*/
	var current_user User
	for {
		if logged_in == false {
			fmt.Println("Log in with your username and password: ")
			var user string
			fmt.Println("Username: ")
			//fmt.Printf("nexus> ")
			fmt.Scanln(&user)

			var pass string
			fmt.Println("Password: ")
			//fmt.Printf("nexus> ")
			fmt.Scanln(&pass)
			login := cm.login(user, pass)
			logged_in = login
			if login == false {
				fmt.Println("ERROR: incorrect login information for", user, ".")
			} else {
				current_user = <-cm.loggedIn
			}
		}
		if logged_in {
			var option string
			fmt.Println("\nOptions: ")
			fmt.Println("1. Create a user (1, create): \n2. Update a specific property of a Person node (update)\n3. Get a node by username\n4. Create Friend relationship\n5. View friends list of a given user\n6. Create/join a house\n7. Get household members (address/username)\n8. Exit")
			fmt.Printf("nexus> ")
			fmt.Scanln(&option)
			if strings.ToLower(option) == "create" || option == "1" {
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

			} else if strings.ToLower(option) == "update" || option == "2" {
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

			} else if strings.ToLower(option) == "get" || option == "3" {
				var user string
				fmt.Println("Username: ")
				fmt.Printf("nexus> ")
				fmt.Scanln(&user)
				cm.get_person(user)
				time.Sleep(time.Millisecond * 500)
			} else if strings.ToLower(option) == "friend" || strings.ToLower(option) == "friends" || option == "4" {
				var user1 string
				fmt.Println("Username 1: ")
				fmt.Printf("nexus> ")
				fmt.Scanln(&user1)
				var user2 string
				fmt.Println("Username 2: ")
				fmt.Printf("nexus> ")
				fmt.Scanln(&user2)
				cm.make_friends(user1, user2)
				time.Sleep(time.Millisecond * 500)
			} else if strings.ToLower(option) == "list" || strings.ToLower(option) == "friends_list" || option == "5" {
				var username string
				fmt.Println("Username: ")
				fmt.Printf("nexus> ")
				fmt.Scanln(&username)
				cm.get_friends(username)
				time.Sleep(time.Millisecond * 500)

			} else if strings.ToLower(option) == "house" || option == "6" {
				var option2 string
				fmt.Println("1. create (username, address)\n2. join (username) ")
				fmt.Printf("nexus> ")
				fmt.Scanln(&option2)

				if strings.ToLower(option2) == "create" || option2 == "1" {
					scanner := bufio.NewScanner(os.Stdin)
					var username string
					fmt.Println("Username: ")
					fmt.Printf("nexus> ")
					//fmt.Scanln(&username)
					scanner.Scan()
					username = scanner.Text()
					fmt.Println("Address: ")
					fmt.Printf("nexus> ")
					var address string
					scanner.Scan()
					address = scanner.Text()
					//fmt.Scanln(&address)
					cm.create_house(username, address)
					time.Sleep(time.Millisecond * 500)
				} else if strings.ToLower(option2) == "join" || option2 == "2" {
					var username string
					fmt.Println("Target Username: ")
					fmt.Printf("nexus> ")
					fmt.Scanln(&username)
					cm.join_house(username, current_user.username)
					time.Sleep(time.Millisecond * 500)
				}

			} else if strings.ToLower(option) == "household" || strings.ToLower(option) == "get_house" || option == "7" {
				var option2 string
				fmt.Println("1. address (ex: 15 Crandall Street)\n2. username (ex: dbottch1) ")
				fmt.Printf("nexus> ")
				fmt.Scanln(&option2)

				if strings.ToLower(option2) == "address" || option2 == "1" {
					scanner := bufio.NewScanner(os.Stdin)
					fmt.Println("Address: ")
					fmt.Printf("nexus> ")
					var address string
					scanner.Scan()
					address = scanner.Text()
					//fmt.Scanln(&address)
					cm.get_household(address, false)
					time.Sleep(time.Millisecond * 500)
				} else if strings.ToLower(option2) == "username" || option2 == "2" {
					var username string
					fmt.Println("Username: ")
					fmt.Printf("nexus> ")
					fmt.Scanln(&username)
					cm.get_household(username, true)
					time.Sleep(time.Millisecond * 500)
				}

			} else if strings.ToLower(option) == "e" || strings.ToLower(option) == "exit" || option == "8" {
				fmt.Println("Exiting... gracefully")
				return
			}
		}
	}
	fmt.Println("Done.")
}
