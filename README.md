# nexus

### Constraints:
If you wish to set up Nexus on your own, you must first install:

* Golang
* Neo4J Community Version (Command Line or Desktop)

Upon configuring and running Neo4J for the first time, as well as creating the database instance,
the first Cypher query you should run is:
```
create constraint uniqueUser ON (n:Person) ASSERT n.username IS UNIQUE
```

This will ensure that each registered user will have a unique username attributed to them.

### Registration: Sign up coming soon!
![Working Image of Nexus Registration Page](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/working%20image%20of%20nexus.PNG)

When a user request to register with Nexus, the following query is sent by the Go driver to the Nexus database to sign them up:
```
CREATE (n:Person { first_name: $first_name, last_name: $last_name, username: $username, password: $password}) RETURN n.first_name, n.last_name, n.username, n.password
```

When a user registers successfully (i.e. their username isn't already taken), a Person node representing them is added to the graph:
![Neo4J Desktop: registering a user/creating a Person node](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/neo4j_create.PNG)

### Login:
![Working Image of Nexus Login Page](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/working%20image%20of%20nexus.PNG)
When a user has created their account, they can then choose to login using their chosen username and password.
The following query is run by the Go driver in order to try and log in:
```
"MATCH (n:Person{username: $username}) return n.first_name, n.last_name, n.username, n.password
```

On the driver side, the password stored in the Neo4J database is compared against the user input:
```go
pw := result.Record().GetByIndex(3).(string)
if login.password == pw {
    cm.loginGood <- true
    user := &User{fn, ln, un, pw}
    //print_user_info(*user)
    cm.loggedIn <- *user
} else {
    cm.loginGood <- false
}
```

### User Portal:
Coming soon...

Backend Queries:
* Getting user information:
If we wanted to get all the information for say, user **dbottch1**, the following query is run with the target username:
```
match (n:Person {username: $u_name}) return n.first_name, n.last_name, n.username, n.password
```
* Update property
If a user were to update their password, the following query would reflect that change in the database:
```
MATCH (n:Person {username: $u_name}) SET $property = $value RETURN $property
```
**Result:**
![Updating a user's password](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/neo4j_update.PNG)
* Add friend (ex: **Let me add you as a friend**):
Let's say that we have two users, Dan and Joe, who wish to be friends:
![Two distinct users](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/neo4j_two.PNG)
```
MATCH (a:Person{username: $u_name1}) , (b:Person{username: $u_name2}) CREATE (a)-[r:FRIEND]->(b) RETURN r, a.username, b.username
```
**Result:**
![Two distinct users becoming friends](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/neo4j_friends.PNG)
* View friends list (ex: **Who are my friends?**):
Imagine more of Dan's friends sign up for Nexus and some of them happen to live in their own house:
![Two distinct users becoming friends](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/many_friends.PNG)

The friends list query grabs all of Dan's friends and displays their information:
```
match (n:Person{username: $username})-[r:FRIEND]->(f) return f.first_name, f.last_name, f.username
```
**Result:**
![Two distinct users becoming friends](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/friends_list.PNG)
* Create House (ex: **Let me add my home address**):
Say Dan (**dbottch1**) wanted to add their home address. A new House node with the unique address is created, and a relationship
between the Person node and the House node is created (**(user:Person)-[:HOUSE]->(h:House)**).
```
match (n:Person{username: $un}) create (h:House{address: $addr}) create (n)-[r:HOUSE]->(h) return h.address
```
**Result:**
![User creating and joining their house](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/create_house.PNG)
* Join House by username (ex: **Let me join my home address**):
Now, say Joe (**jsanch49**) wanted to join dbottch1's house, as they live together:
```
match (n:Person{username: $target_user})-[r:HOUSE]->(h), (c:Person{username: $current_user}) create (c)-[:HOUSE]->(h) return h.address
```
**Result**:
![User creating and joining their house](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/join_house.PNG)
* View household members:
By address (ex: **Who lives at 15 Crandall?**):
```
match (h:House{address: $input})<--(n) return n.username, n.first_name, n.last_name
```
**Result**:
![User creating and joining their house](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/housemates_addr.PNG)
By username (ex: **Who lives with 'mperaza'?**):
```
match (h:House)<-[r:HOUSE]-(n:Person{username: $input}) match (h)<-[:HOUSE]-(p) return p.username, p.first_name, p.last_name
```
**Result**:
![User creating and joining their house](https://github.com/hctob/nexus/blob/main/nexus-frontend/img/housemates_username.PNG)
