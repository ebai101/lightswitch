package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	args := os.Args
	baseUrl := "http://192.168.1.105:3000/"
	client := &http.Client{}

	var command string
	if len(args) < 2 {
		command = "get"
	} else {
		command = args[1]
	}

	if command == "get" {
		resp, err := client.Get(baseUrl)
		if err != nil {
			log.Fatal(err)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if len(body) > 0 {
			fmt.Println(string(body))
		}

		defer resp.Body.Close()
	} else {
		resp, err := client.PostForm(baseUrl+"?"+command, nil)
		if err != nil {
			log.Fatal("invalid light command:", command)
		}

		defer resp.Body.Close()
	}
}
