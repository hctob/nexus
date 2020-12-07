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
    arg_username_raw     = flag.String("username", "dan", "Usernames are unique identifiers for database users.")
    arg_password_raw     = flag.String("password", "test", "Unencrypted password for selected username.")

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

	greeting, err := session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
            "CREATE (n:Person { first_name: $fn, last_name: $ln }) RETURN n.first_name, n.last_name", map[string]interface{}{
            "first_name":   "Quindarius",
            "last_name": "Gooch",
            })
		if err != nil {
			return nil, err
		}
        fmt.Println("node creation query passed\n")
		if result.Next() {
			return result.Record().GetByIndex(0), nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return "", err
	}

	return greeting.(string), nil
}

/*
result, err := session.Run("CREATE (n:Person { first_name: $fn, last_name: $ln }) RETURN n.first_name, n.last_name", map[string]interface{}{
"first_name":   "Quindarius",
"last_name": "Gooch",
})
*/

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

    ret, _ := helloWorld("bolt://localhost:7687", *arg_password_raw, *arg_password_raw, false)
    fmt.Println("Done: ", ret)
}
