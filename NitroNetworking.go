package main

import (
    "fmt"
    "net/https"
    "log"
    "html/template"
    "github.com/gorilla/mux"
    "github.com/google/gopacket/pcap"
)

type App struct {
    Title string
    Author string
}
func main() {
    router := mux.NewRouter().StrictSlashes(true)
    router.HandleFunc("/",IndexHandler)
    router.HandleFunc("/",NetworkHandler)

    log.Fatal(http.ListenAndServe(":8080",router))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    //Load index template
    t, err := template.ParseFile("html/index.html")
    if err != nil {
        fmt.Fprintf(w, "Error parsing template file.")
    }
    info := App{Title:"Rascal Networking", Author: "David Piedra"}
    t.exec(w, info)
}