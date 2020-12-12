package main

/*import (
    "fmt"
    //"log"
    "html/template"
    "encoding/json"
	"net/http"
)

type JSON_User struct {
	Firstname string   `json:firstname`
	Lastname  string   `json:lastname`
	Username   string  `json:username`
	Password   string  `json:password`
	Atrisk		string `json:atrisk`
    Lastdate    string `json:lastdate`
}

func login_page(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Error(w, "404 not found.", http.StatusNotFound)
        return
    }
    switch r.Method {
    case "GET":
         http.ServeFile(w, r, "nexus-frontend/login.html")
    case "POST":
        // Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
        if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
        }
        //fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
        username := r.FormValue("username")
        pw := r.FormValue("password")

        fmt.Printf("Client-side Username = %s\n", username)
        fmt.Printf("Client-side password = %s\n", pw)
        //http.ServeFile(w, r, "user.html")
        good := (*cm_global).login(username, pw)
        if good {
            current_user = <-(*cm_global).loggedIn
            http.Redirect(w, r, "nexus-frontend/user/user.html", http.StatusFound)
        }
    default:
        fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
    }
}

func userHandler(w http.ResponseWriter, r *http.Request) {
      w.Header().Set("Content-Type", "text/html; charset=utf-8")

      t, err := template.ParseFiles("nexus-frontend/user/user.html")
      if err != nil {
             fmt.Fprintf(w, "Unable to load template")
        }



    user := &JSON_User{current_user.first_name, current_user.last_name, current_user.username, current_user.password, current_user.at_risk, current_user.last_infected_time}

     t.Execute(w, *user)
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
      w.Header().Set("Content-Type", "application/json")
      user := JSON_User{
                    Firstname: "Dan",
                    Lastname: "Bottcher",
                    Username: "dbottch1",
                    Password: "test",
                    Atrisk: "false",
                    Lastdate: "N/A",
                 }

     json.NewEncoder(w).Encode(user)
}

/*func main() {
    http.HandleFunc("/", login_page)
    http.HandleFunc("/json/", jsonHandler)
    http.HandleFunc("/user/", userHandler)
    http.HandleFunc("/test/", testHandler)
    http.HandleFunc("/main/", mainHandler)
    http.HandleFunc("/friends_list/", listHandler)
    http.HandleFunc("/add-friend/", friendHandler)
    fmt.Printf("Starting server for testing HTTP POST...\n")
    if err := http.ListenAndServe(":9001", nil); err != nil {
        log.Fatal(err)
    }
}*/
