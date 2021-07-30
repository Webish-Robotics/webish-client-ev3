package main

import (
	"log"
	"runtime"
	"bytes"
	"io/ioutil"
	"os/exec"
	"encoding/json"

	socket "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

type Architecture struct {
	Code string `json:"code"`
	User string `json:"user"`
	Room string `json:"room"`
}

type Feedback struct {
	Message string `json:"message"`
	Username string `json:"username"`
	Room string `json:"room"`
	Code int `json:"code"`
}

const found string = "found"
const occupied string = "occupied"

var process int = 0
var clientHandler *socket.Client

const exit_signal int = 1
const signal_killed int = 2
const success_code int = 0

func writeScript(code string) {
	octets := []byte(code)

	err := ioutil.WriteFile("app.py", octets, 0644)
	if err != nil {
		log.Fatal("Couldn't execute the writeScript")
	}
}

func launchScript(arch Architecture) {
	writeScript(arch.Code)
	cmd := exec.Command("python3", "./app.py")

	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err := cmd.Start()
	process = cmd.Process.Pid

	err = cmd.Wait()
	code := success_code

	if err != nil {
		if err.Error() == "exit status 1" {
			code = exit_signal
		}
		if err.Error() == "signal: killed" {
			code = signal_killed
		}
		errStr := string(stderr.Bytes())

		bErr, _ := json.Marshal(Feedback{errStr, arch.User, arch.Room + "STUDENT", code})
		clientHandler.Emit("/feedback", string(bErr))
	} else {
		outStr := string(stdout.Bytes())
		oErr, _ := json.Marshal(Feedback{outStr, arch.User, arch.Room + "STUDENT", code})
		clientHandler.Emit("/feedback", string(oErr))
	}
	process = 0
}

func resetRobot() {
	cmd := exec.Command("bash", "reset.sh")
	err := cmd.Run()

	if err != nil {
		log.Println(" => ERROR with client at reset.sh:", err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	c, err := socket.Dial(socket.GetUrl("169.254.27.203", 8080, false), transport.GetDefaultWebsocketTransport())

	if err != nil {
		log.Fatal(err)
		c.Close()
	}

	err = c.On(socket.OnDisconnection, func(conn *socket.Channel) {
		log.Println("OnDisconnect callback was reached.")
	})
	if err != nil {
		log.Fatal(err)
		c.Close()
	}

	err = c.On(socket.OnConnection, func(conn *socket.Channel) {
		log.Println("OnConnection callback was reached.")
		room, err := ioutil.ReadFile("room.txt")
		if err != nil {
			log.Fatal(err)
		}

		c.Emit("/joinable", string(room))
	})
	if err != nil {
		log.Fatal(err)
		c.Close()
	}

	clientHandler = c
	for {
		err = c.On("/check", func(conn *socket.Channel, args interface{}) string {
			ret := found
			if process != 0 {
				ret = occupied
			}
			return ret
		})
		err = c.On("/serve", func(conn *socket.Channel, arch Architecture) {
			go launchScript(arch)
		})

		if err != nil {
			log.Fatal(err)
		}
	}
}
