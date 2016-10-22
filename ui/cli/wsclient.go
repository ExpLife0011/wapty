package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"

	"github.com/empijei/Wapty/intercept"
	"github.com/empijei/Wapty/ui"
	"golang.org/x/net/websocket"
)

var ws *websocket.Conn
var serverChannel chan ui.Command
var stdin *bufio.Scanner

func init() {
	serverChannel = make(chan ui.Command)
	stdin = bufio.NewScanner(os.Stdin)
}

func main() {
	go wsLoop()
	cli()
}

func cli() {
	for cmd := range serverChannel {
		_ = ioutil.WriteFile("tmp.swp", *cmd.Payload, 0644)
		log.Println("Payload intercepted, edit it and press enter to continue.")
		var payload []byte
		var args ui.Args
		for args == nil {
			stdin.Scan()
			switch stdin.Text() {
			case intercept.EDITED.String():
				payload, _ = ioutil.ReadFile("tmp.swp") //TODO chech this error
				args = ui.Args(map[string]string{"action": intercept.EDITED.String()})
			case intercept.FORWARDED.String():
				args = ui.Args(map[string]string{"action": intercept.FORWARDED.String()})
			default:
				log.Println("Unknown command")
				log.Println("Try with ", intercept.EDITED, " ", intercept.FORWARDED)
			}
		}
		log.Println("Continued")
		err := websocket.JSON.Send(ws, ui.Command{Args: args, Channel: intercept.EDITORCHANNEL, Payload: &payload})
		if err != nil {
			panic(err)
		}
	}
}

func wsLoop() {
	var url = "ws://localhost:8081/ws"
	var origin = "http://localhost/"
	var err error
	ws, err = websocket.Dial(url, "", origin)
	if err != nil {
		panic(err)
	}
	for {
		var msg ui.Command
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			panic(err)
		}
		serverChannel <- msg
	}

}
