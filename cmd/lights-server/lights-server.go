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

type Presets map[string]map[string]string

var presets Presets
var state = map[string]string{"bed": "0,0,0,0,0", "left": "0,0,0,0,0", "right": "0,0,0,0,0", "back": "0,0,0,0,0"}
var client mqtt.Client

func readPresetFile(presetFilePath string, presetData *Presets) {
	file, err := os.Open(presetFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileBytes, _ := io.ReadAll(file)
	json.Unmarshal(fileBytes, &presetData)
}

func setLight(lightName string, lightColor string) {
	if token := client.Publish(fmt.Sprintf("light/cmnd/%s/color", lightName), 0, false, lightColor); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}
}

func setPreset(preset map[string]string) {
	for k, v := range preset {
		setLight(k, v)
	}
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		io.WriteString(w, fmt.Sprintf("%v\n", state))
	case "POST":
		presetName := r.URL.Query().Get("preset")
		if preset, ok := presets[presetName]; ok {
			setPreset(preset)
		}

	default:
		panic(fmt.Sprintf("got strange request method: %s", r.Method))
	}
}

var pubHandler = func(client mqtt.Client, msg mqtt.Message) {
	var statResult map[string]string
	json.Unmarshal(msg.Payload(), &statResult)

	r := regexp.MustCompile(`stat/(.*)/RESULT`)
	lightName := r.FindStringSubmatch(msg.Topic())[1]
	state[lightName] = statResult["Color"]
	fmt.Println(lightName, state[lightName])
}

func main() {
	args := os.Args
	var presetFilePath string
	if len(args) < 2 {
		log.Fatal("no presets provided, pass them as a cli arg")
	} else {
		presetFilePath = args[1]
	}
	readPresetFile(presetFilePath, &presets)

	mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker("mqtt://localhost:1883")
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
