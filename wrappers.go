package main

func get_contact_list(s *Session) ([]person){
  result, _ := session.Run(`
        MATCH (p:Person) MATCH (p)-[*]-() RETURN n`, nil)

  for result.Next() {
      record := result.Record()
      nodeID, ok := record.Get("nodeId")
      // ok == true and nodeID is an interface that can be asserted to int
      username, ok := record.Get("username")
      // ok == true and username is an interface that can be asserted to string

  }
}

//connects two people
func create_contact_person(u_name1 string, u_name2 string, s *Session){
  result, _ := (*s).Run(`
        MATCH (a:Person),(b:Person)
WHERE a.username = '$u_name1' AND b.username = '$u_name2'
CREATE (a)-[r:Friend]->(b) RETURN r`, nil)
}

//connects a person to a house
func create_contact_house(u_name string, address string, s *Session){
  result, _ := (*s).Run(`
        MATCH (a:Person),(h:House)
WHERE a.username = '$u_name' AND h.address = '$address'
CREATE (a)-[r:House]->(b) RETURN r`, nil)
}

//update a user's property
func update_user(u_name string, property string, value string, s *Session){
  result, _ := (*s).Run(`
        MATCH (a:Person)
WHERE a.username = '$u_name' SET a.$property = $value RETURN a`, nil)
}
