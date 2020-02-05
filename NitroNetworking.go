package main

import (
    "fmt"
    "net/https"
    "log"
    "html/template"
    "github.com/gorilla/mux"
    "github.com/google/gopacket/pcap"
)
type DeviceList struct {
    Content string
    DeviceCount int
}
type App struct {
    Title string
    Author string
}
func main() {
    router := mux.NewRouter().StrictSlashes(true)
    router.HandleFunc("/",IndexHandler)
    router.HandleFunc("/network",NetworkHandler)

    log.Fatal(http.ListenAndServe(":8080",router))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    //Load index template
    t, err := template.ParseFile("html/index.html")
    if err != nil {
        fmt.Fprintf(w, "Error parsing template file.")
    }
    info := App{Title:"Rascal Networking", Author: "David Piedra"}
    t.Execute(w, info)
}

func NetworkHandler(w http.ResponseWriter, r *http.Request) {
    t, err := template.ParseFile("html/network.html")
    if err != nil {
        fmt.Fprintf(w, "Error parsing network file.")
    }
    content := ListDevices()
    
    t.Execute(w, content)
}

func ListDevices() DeviceList {
    dList := "<div class='dev-list'>\n"
    //Use pcap to get all devices on the network
    devices, err := pcap.FindAllDevs()
    if err != nil {
        log.Fatal(err)
    }
    count := 0
    for _, device := range devices {
        dList += "<div class='dev-desc'>\n" + device.Description + "\n</div>"
        dList += "<div class='dev-name'>\n" + device.Name + "\n</div>"
        dList += "<ul class='dev-addresses'>\n"
        for _, address := range device.Addresses {
            dList += "<li>IP: " + address.IP.String() + " | Subnet mask: " + address.Netmask.String() + "</li>"
        }
        dList += "</ul>"
        count++
    }
    dList += "</div>\n"

    finalList := DeviceList{Content: dList, DeviceCount: count}
    return &finalList
}