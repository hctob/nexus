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
var (
	arg_uri     = flag.String("uri", "bolt://localhost:7687", "The URI for the Nexus database, to connect to it.")
    arg_username_raw     = flag.String("u", "dan", "Usernames are unique identifiers for database users.")
    arg_password_raw     = flag.String("p", "test", "Unencrypted password for selected username.")

    totalQueries int64
)

func helloWorld(uri, username, password string, encrypted bool) (string, error) {
	driver, err := neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""), func(c *neo4j.Config) {
		c.Encrypted = encrypted
	})
	if err != nil {
		return "driver connection error", err
	}
    fmt.Println("established driver connection\n")
	defer driver.Close()

	session, err := driver.Session(neo4j.AccessModeWrite)

	if err != nil {
		return "session creation error", err
	}
    fmt.Println("created session with writemode\n")
	defer session.Close()

	/*greeting, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
            "CREATE (n:Person { first_name: $fn, last_name: $ln }) RETURN n.first_name, n.last_name", map[string]interface{}{
            "first_name":   "Quindarius",
            "last_name": "Gooch",
        })*/
        result, err := session.Run("CREATE (n:Person { first_name: $fn, last_name: $ln }) RETURN n.first_name, n.last_name", map[string]interface{}{
        "first_name":   "Quindarius",
        "last_name": "Gooch", })

	if err != nil {
		return "Error: ", err
	}
    /*fmt.Println("node creation query passed\n")
	if result.Next() {
		return result.Record().GetByIndex(0), nil
	}

	return nil, result.Err()
	})*/
	if err != nil {
		return "", result.Err()
	}

	return "done", nil
}

func drive(uri, username, password string) error {
        // configForNeo4j35 := func(conf *neo4j.Config) {}
    configForNeo4j40 := func(conf *neo4j.Config) { conf.Encrypted = false }

    driver, err := neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth(username, password, ""), configForNeo4j40)
    if err != nil {
    	return err
    }
    fmt.Println("established driver connection\n")
    // handle driver lifetime based on your application lifetime requirements
    // driver's lifetime is usually bound by the application lifetime, which usually implies one driver instance per application
    defer driver.Close()

    // For multidatabase support, set sessionConfig.DatabaseName to requested database
    sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
    session, err := driver.NewSession(sessionConfig)
    if err != nil {
    	return err
    }
    fmt.Println("established session connection\n")
    defer session.Close()

    result, err := session.Run("CREATE (n:Person { first_name: $first_name, last_name: $last_name, username: $username, password: $password}) RETURN n.first_name, n.last_name, n.username, n.password", map[string]interface{}{
    "first_name":   "Milton",
    "last_name": "Peraza",
    "username": "mperaza",
    "password": "testing", })

    if err != nil {
    	return err
    }
    fmt.Println("query fine\n")
    for result.Next() {
    	fmt.Printf("Created Person '%s %s' with username = '%s'\n", result.Record().GetByIndex(0).(string), result.Record().GetByIndex(1).(string), result.Record().GetByIndex(2).(string))
    }
    return result.Err()
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

    err := drive("bolt://localhost:7687", *arg_username_raw, *arg_password_raw)
    if err != nil {
		fmt.Println("Error:\n", err)
	}
    fmt.Println("Done.")
}
