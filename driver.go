package main

import (
    "fmt"
    "flag"
    //"time"
    "runtime"
    "github.com/neo4j/neo4j-go-driver/neo4j"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "strconv"
    "time"
    //"html/template"
    //"log"
)


var (
	arg_uri     = flag.String("uri", "bolt://localhost:7687", "The URI for the Nexus database, to connect to it.")
    arg_username_raw     = flag.String("u", "neo4j", "Usernames are unique identifiers for database users.")
    arg_password_raw     = flag.String("p", "cs476", "Unencrypted password for selected username.")
    cm_global * ChannelPool

    logged_in = false
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
            	fmt.Printf("\nCreated Person '%s %s' with username = '%s'\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
            }
            //cm.created <- true
        case house := <-cm.createHouse:
    			un := house.username
    			addr := house.address
    			result, err := session.Run("match (n:Person{username: $un}) create (h:House{address: $addr, should_quarantine: 'false'}) create (n)-[r:HOUSE]->(h) return h.address", map[string]interface{}{
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
    			result, err := session.Run("match (n:Person{username: $username})-[r:FRIEND]->(f) return f.first_name, f.last_name, f.username, f.at_risk, f.last_infected_time order by f.last_infected_time desc limit 50", map[string]interface{}{
    				"username": username,
    			})
    			if err != nil {
    				fmt.Println("Error:\n", err)
    				return
    			}
    			//fmt.Println("query fine\n")
    			fmt.Printf("\nFriends list of %s: \n", username)
    			for result.Next() {
    				fmt.Printf("%s %s: %s, at_risk = %s, last infected time = %s\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string), result.Record().GetByIndex(3).(string), result.Record().GetByIndex(4).(string))
    			}
        case update := <-cm.updateChannel:
            //update_user(update.username, update.property, update.value, &session)
            prop := "n." + update.property
            /*fmt.Println(prop)*/

            v := update.value
            un := update.username
            result, err := session.Run(fmt.Sprintf("MATCH (n:Person {username: $u_name}) SET %s = $value RETURN %s", prop, prop), map[string]interface{}{
              "value": v,
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
              "u_name": username, })

              if err != nil {
                  fmt.Println("Error:\n", err)
                  return
              }
              //fmt.Println("query fine\n")
              for result.Next() {
                  fmt.Printf("\nNode: \"%s %s\":\nusername :%s\npassword: %s\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string), result.Record().GetByIndex(3).(string))
              }

        case friends := <-cm.friendChannel:
            un1 := friends.u_name1
            un2 := friends.u_name2
            fmt.Println(un1, " ", un2)
            result, err := session.Run("MATCH (a:Person{username: $u_name1}) , (b:Person{username: $u_name2}) CREATE (a)-[r:FRIEND]->(b) RETURN r, a.username, b.username", map[string]interface{}{
                "u_name1" : un1,
                 "u_name2" : un2,
            })

            if err != nil {
                fmt.Println("Error:\n", err)
                return
            }
            for result.Next() {
                fmt.Printf("\n%s\ncreated between %s and %s\n", result.Record().GetByIndex(0), result.Record().GetByIndex(1), result.Record().GetByIndex(2))
            }
        case username := <-cm.notify_chan:
    			//fmt.Println("Username: ", username)
    			//match (p)-[:FRIEND]->(f:Person) set f.at_risk = 'true' will also set all friends to true
    			_, err := session.Run("match (h:House)<-[r:HOUSE]-(n:Person{username: $input}) match (h)<-[:HOUSE]-(p:Person) match (p)-[:FRIEND]->(f:Person) set f.at_risk = 'true' set p.at_risk = 'true' set n.at_risk = 'positive' set h.should_quarantine = 'true'", map[string]interface{}{
    				"input": username,
    			})
    			if err != nil {
    				fmt.Println("Error:\n", err)
    				return
    			}
    			/*for result.Next() {
    				fmt.Printf("\nHousemate %s is now at risk!\n", result.Record().GetByIndex(0).(string))
    			}*/
    			fmt.Println("Friends and housemates are now exposed!")
        case test := <-cm.update_test:
    			result, err := session.Run("match (n:Person{username: $username}) set n.at_risk = $result set n.last_infected_time = $date return n.at_risk, n.last_infected_time, n.username", map[string]interface{}{
    				"username": test.username,
    				"result": test.result,
    				"date": test.date,
    			})
    			if err != nil {
    				fmt.Println("Error:\n", err)
    				return
    			}
    			for result.Next() {
    				fmt.Printf("\nUpdated most recent test results for %s:\nresult: %s\tdate: %s\n", result.Record().GetByIndex(2).(string), result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string))
    			}
    		case house_raw := <-cm.createHouseRaw:
    			addr := house_raw.address
    			result, err := session.Run("create (h:House{address: $addr, should_quarantine: 'false'}) return h.address", map[string]interface{}{
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
    		case join_addr := <-cm.joinHouseAddr:
    			result, err := session.Run("match (n:Person{username: $target_user}), (h:House{address: $addr}) create (n)-[:HOUSE]->(h) return n.username, h.address", map[string]interface{}{
    				"target_user":  join_addr.target_user,
    				"addr": join_addr.address,
    			})
    			if err != nil {
    				fmt.Println("Error:\n", err)
    				return
    			}
    			//fmt.Println("query fine\n")
    			for result.Next() {
    				fmt.Printf("\n%s joined House with address %s\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string))
    			}
      case login := <-cm.loginChannel:
          //time.Sleep(20 * time.Second)
          un := login.username
          //fmt.Println("Received username =", un)
          pw := login.password
          //fmt.Println("Received password =", pw)
          result, err := session.Run("MATCH (a:Person{username: $username}) return a.password", map[string]interface{}{
              "username" : un,
               "password" : pw,
          })

        if err != nil {
            fmt.Println("Error:\n", err)
            return
        }

        //res := false
        for result.Next() {
            fmt.Println(login.password, result.Record().GetByIndex(0).(string))
            if login.password == result.Record().GetByIndex(0).(string) {
                cm.loginGood <- true
                fmt.Println("successful login\n")
            } else {
                cm.loginGood <- false
            }
        }
        //fmt.Println("successful login\n")
        //cm.loginGood <- false
        }
    }
}

type Login_Frontend struct {
    Username string
    Login_status bool
}
type Login_Response struct {
  Username string `json: Username`
  Status bool `json: Status`
}

func get_login(w http.ResponseWriter, req *http.Request){
  if req.Method == http.MethodPost{
    var obj map[string]string
    body, readErr := ioutil.ReadAll(req.Body)
    if readErr != nil{
      panic(readErr)
    }
    err := json.Unmarshal(body, &obj)
    if err != nil{
      panic(err)
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    fmt.Println("username: ", obj["Username"], "\n")
    fmt.Println("password: ", obj["Password"], "\n")
    t := (*cm_global).login(obj["Username"], obj["Password"])
    fmt.Println("working")

    cookie := http.Cookie{Name: "Username", Value: string(obj["Username"])}
    http.SetCookie(w, &cookie)
    cookie_status := http.Cookie{Name: "Status", Value: strconv.FormatBool(t)}
    http.SetCookie(w, &cookie_status)
    info := Login_Response{
      Username: obj["Username"],
      Status: t,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(info)
  }
}

func add_friend(w http.ResponseWriter, req *http.Request){
  if req.Method == http.MethodPost{
    var obj map[string]string
    body, readErr := ioutil.ReadAll(req.Body)
    if readErr != nil{
      panic(readErr)
    }
    err := json.Unmarshal(body, &obj)
    if err != nil{
      panic(err)
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    fmt.Println(obj["Username"])
    fmt.Println(obj["Friend_username"])
    (*cm_global).make_friends(obj["Username"], obj["Friend_username"])
  }
}

func update_test(w http.ResponseWriter, req *http.Request){
  if req.Method == http.MethodPost{
    var obj map[string]string
    body, readErr := ioutil.ReadAll(req.Body)
    if readErr != nil{
      panic(readErr)
    }
    err := json.Unmarshal(body, &obj)
    if err != nil{
      panic(err)
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    fmt.Println(obj["Result"])
    fmt.Println(obj["Date"])
    fmt.Println(obj["Username"])
    now := time.Now()
    sent, err := time.Parse("2006-01-02", obj["Date"])
    if err != nil{
      panic(err)
    }
    newest_result, _ := strconv.ParseBool(obj["Result"])
    if(newest_result){
      diff := now.Sub(sent).Hours() / 24
      if(diff < 5){
        (*cm_global).add_test("true", obj["Date"], obj["Username"])
      }else{
        (*cm_global).add_test("false", obj["Date"], obj["Username"])
      }
    }else{
      (*cm_global).add_test("false", obj["Date"], obj["Username"])
    }
  }
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
    /*cm.createChannel = make(chan User, 128)
    cm.created = make(chan bool)
    cm.updateChannel = make(chan Update, 128)
    cm.getNodeChannel = make(chan string, 128)
    */

    cm = pool_init()
    cm_global = &cm

    //cm.create_person("test", "test", "test", "test")

    go drive("bolt://localhost:7687", *arg_username_raw, *arg_password_raw, cm)

    //cm.create_person("Will", "Kennedy", "neo4j", "neo4j")
    http.Handle("/", http.FileServer(http.Dir("./nexus-frontend")))
    http.HandleFunc("/add-friend", add_friend)
    http.HandleFunc("/update-test", update_test)
    http.HandleFunc("/login", get_login)
    http.ListenAndServe(":8090", nil)

    /*for {
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
            login := cm.login(user,pass)
            logged_in = login
            if login == false {
                fmt.Println("ERROR: incorrect login information for", user, ".")
            }
        }
        if logged_in {
            var option string
            fmt.Println("\nOptions: ")
            fmt.Println("1. Create a user (1, create): \n2. Update a specific property of a Person node (update)\n3. Get a node by username\n4. Create Friend relationship \n5. Exit")
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

            } else if option == "get" || option == "3" {
                var user string
                fmt.Println("Username: ")
                fmt.Printf("nexus> ")
                fmt.Scanln(&user)
                cm.get_person(user)
                time.Sleep(time.Millisecond * 500)
            } else if option == "friend" || option == "friends" || option == "4" {
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
            } else if option == "e" || option == "exit" || option == "5" {
                fmt.Println("Exiting... gracefully")
                return
            }
    }*/
//}
    fmt.Println("Done.")
}
