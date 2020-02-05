package main

import (
    "fmt"
    "net/https"
    "net"
    "log"
    "html/template"
    "github.com/gorilla/mux"
    "github.com/google/gopacket/pcap"
    "github.com/muka/go-bluetooth"
)
type DeviceList struct {
    Content string
    DeviceCount int
}
type App struct {
    Title string
    Author string
}
//Global Variables
var BluetoothDevices = ""

func main() {
    router := mux.NewRouter().StrictSlashes(true)
    router.HandleFunc("/",IndexHandler)
    router.HandleFunc("/network",NetworkHandler)
    router.HandleFunc("/bluetooth",BluetoothHandler)

    log.Fatal(http.ListenAndServe(":8080",router))
}
//All Page Handler's here
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

func BluetoothHandler(w http.ResponseWriter, r *http.Request) {
    t, err := template.ParseFile("html/bluetooth.html")
    if err != nil {
        fmt.Fprintf(w, "Error parsing bluetooth file.")
    }
    
    content := ListBluetoothDevices()
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
func ListBluetoothDevices() DeviceList {
    


}


//Helper functions
func Run(adapterID string, onlyBeacon bool) error {
    //https://github.com/muka/go-bluetooth/blob/master/examples/discovery/discovery.go <- helped
	//clean up connection on exit
	defer api.Exit()

	a, err := adapter.GetAdapter(adapterID)
	if err != nil {
		return err
	}

	log.Debug("Flush cached devices")
	err = a.FlushDevices()
	if err != nil {
		return err
	}

	log.Debug("Start discovery")
	discovery, cancel, err := api.Discover(a, nil)
	if err != nil {
		return err
	}
	defer cancel()

	go func() {

		for ev := range discovery {
            action := 0
			if ev.Type == adapter.DeviceRemoved {
                action = 1
				continue
			}

			dev, err := device.NewDevice1(ev.Path)
			if err != nil {
				log.Errorf("%s: %s", ev.Path, err)
				continue
			}

			if dev == nil {
				log.Errorf("%s: not found", ev.Path)
				continue
			}

			log.Infof("name=%s addr=%s rssi=%d", dev.Properties.Name, dev.Properties.Address, dev.Properties.RSSI)

			content, err = handleBeacon(dev)
			if err != nil {
				log.Errorf("%s: %s", ev.Path, err)
            }
            else if err == nil && content == "" {
                log.Error("Was not a beacon")
            } else {
                if action == 1 {
                    //TODO: Device Removed, take it out of string
                    stringToSearchFor := "<li>" + content + "</li>\n"
                    BluetoothDevices = strings.Replace(BluetoothDevices, stringToSearchFor, "", 1)
                } else {
                    BluetoothDevices += "<li>" + content + "</li>\n"
                }
            }

		}

	}()

	select {}
}
func handleBeacon(dev *device.Device1) (string, error) {
    b, err := beacon.NewBeacon(dev)
	if err != nil {
		return "",err
	}

	beaconUpdated, err := b.WatchDeviceChanges(context.Background())
	if err != nil {
		return "",err
	}

	isBeacon := <-beaconUpdated
	if !isBeacon {
		return "", nil
	}

	name := b.Device.Properties.Alias
	if name == "" {
		name = b.Device.Properties.Name
	}

	log.Debugf("Found beacon %s %s", b.Type, name)
    content := "Type: " + b.Type + " | Name: " + name
	if b.IsEddystone() {
		eddystone := b.GetEddystone()
		switch eddystone.Frame {
		case beacon.EddystoneFrameUID:
			log.Debugf(
				"Eddystone UID %s instance %s (%ddbi)",
				eddystone.UID,
				eddystone.InstanceUID,
				eddystone.CalibratedTxPower,
			)
			break
		case beacon.EddystoneFrameTLM:
			log.Debugf(
				"Eddystone TLM temp:%.0f batt:%d last reboot:%d advertising pdu:%d (%ddbi)",
				eddystone.TLMTemperature,
				eddystone.TLMBatteryVoltage,
				eddystone.TLMLastRebootedTime,
				eddystone.TLMAdvertisingPDU,
				eddystone.CalibratedTxPower,
			)
			break
		case beacon.EddystoneFrameURL:
			log.Debugf(
				"Eddystone URL %s (%ddbi)",
				eddystone.URL,
				eddystone.CalibratedTxPower,
			)
			break
		}

	}
	if b.IsIBeacon() {
		ibeacon := b.GetIBeacon()
		log.Debugf(
			"IBeacon %s (%ddbi) (major=%d minor=%d)",
			ibeacon.ProximityUUID,
			ibeacon.MeasuredPower,
			ibeacon.Major,
			ibeacon.Minor,
		)
	}

	return content, nil
}
func getLocalIP() net.UDPAddr {
    conn, err := net.Dial("udp","8.8.8.8:80")
    if err != nil {
        log.Fatal("Couldn't start outbound connection")
    }
    defer conn.Close()
    localAddr := conn.LocalAddr().(*net.UDPAddr)
    return &localAddr
}

