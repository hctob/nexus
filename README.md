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
