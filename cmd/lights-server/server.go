package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var topics = []string{"back", "left", "right", "bed"}
var state = map[string]string{"bed": "0,0,0,0,0", "left": "0,0,0,0,0", "right": "0,0,0,0,0", "back": "0,0,0,0,0"}
var presets = map[string]map[string]string{
	"off":  {"bed": "0,0,0,0,0", "left": "0,0,0,0,0", "right": "0,0,0,0,0", "back": "0,0,0,0,0"},
	"day":  {"bed": "0,0,0,0,255", "left": "0,0,0,0,255", "right": "0,0,0,0,255", "back": "0,0,0,0,255"},
	"bed":  {"bed": "100,0,0,0,50", "left": "0,0,0,0,0", "right": "0,0,0,0,0", "back": "0,0,0,0,0"},
	"eve":  {"bed": "70,0,0,0,75", "left": "50,0,0,0,112", "right": "0,0,0,0,135", "back": "0,0,0,0,0"},
	"nite": {"bed": "100,0,0,0,50", "left": "0,100,100,50,0", "right": "0,90,125,0,0", "back": "0,0,0,0,0"},
}
var client mqtt.Client

func setLight(lightName string, lightColor string) {
	if token := client.Publish(fmt.Sprintf("light/cmnd/%s/color", lightName), 0, false, lightColor); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}
}

func setPreset(preset string) {
	for _, topic := range topics {
		setLight(topic, presets[preset][topic])
	}
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		io.WriteString(w, "hello this is not implemented lmao\n")
	case "POST":
		preset := r.URL.Query().Get("preset")
		if preset != "" {
			setPreset(preset)
		}

	default:
		panic(fmt.Sprintf("got strange request method: %s", r.Method))
	}
}

var pubHandler = func(client mqtt.Client, msg mqtt.Message) {
	// handler
	var statResult map[string]string
	json.Unmarshal(msg.Payload(), &statResult)

	r := regexp.MustCompile(`stat/(.*)/RESULT`)
	lightName := r.FindStringSubmatch(msg.Topic())[1]
	state[lightName] = statResult["Color"]
	fmt.Println(lightName, state[lightName])
}

func main() {
	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker("mqtt://192.168.1.105:1883")
	opts.SetClientID("lightserver")
	opts.SetKeepAlive(2 * time.Second)
	opts.SetDefaultPublishHandler(pubHandler)
	opts.SetPingTimeout(1 * time.Second)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("mqtt connected")

	if token := client.Subscribe("light/stat/#", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	http.HandleFunc("/", reqHandler)
	http.ListenAndServe(":3000", nil)
}
