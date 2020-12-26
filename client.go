package main

import (
	"log"
	"runtime"
	"bytes"
	"io/ioutil"
	"os/exec"
	"fmt"

	socket "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

type Architecture struct {
	Code string `json:"code"`
	User int `json:"user"`
	Room string `json:"room"`
}

type Serve struct {
	Code string `json:"code"`
	User int `json:"user"`
	SocketUser *socket.Channel `json:"socketuser"`
}

type Robot struct {
	Process int `json:"process"`
	Result string `json:"result"`
	ResultID int `json:"resultid"`
	Client *socket.Channel `json:"user"`
}

const found string = "found"
const occupied string = "occupied"

const exit_signal int = 1
const signal_killed int = 2
const success_code int = 0

var framework Robot = Robot{0, "", 0, nil}

func writeScript(code string) {
	octets := []byte(code)

	err := ioutil.WriteFile("app.py", octets, 0644)
	if err != nil {
		log.Fatal("Couldn't execute the writeScript")
	}
}

func launchScript(code string) string {
	writeScript(code)
	cmd := exec.Command("python3", "./app.py")

	var stdout, stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	err := cmd.Start()
	framework.Process = cmd.Process.Pid

	err = cmd.Wait()
	if err != nil {
		if err.Error() == "exit status 1" {
			framework.ResultID = exit_signal
		}
		if err.Error() == "signal: killed" {
			framework.ResultID = signal_killed
		}
		errStr := string(stderr.Bytes())
		framework.Result = errStr
	} else {
		outStr := string(stdout.Bytes())
		framework.ResultID = success_code
		framework.Result = outStr
	}
	return framework.Result
}

func resetRobot() {
	cmd := exec.Command("bash", "/client/reset.sh")
	err := cmd.Run()

	if err != nil {
		fmt.Println(" => ERROR with client at /client/reset.sh:", err)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	c, err := socket.Dial(socket.GetUrl("169.254.247.202", 8080, false), transport.GetDefaultWebsocketTransport())
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
		c.Emit("/joinable", "demo")
	})
	if err != nil {
		log.Fatal(err)
		c.Close()
	}

	for {
		err = c.On("/check", func(conn *socket.Channel, args interface{}) string {
			ret := found
			if framework.Process != 0 {
				ret = occupied
			}
			return ret
		})
		err = c.On("/serve", func(conn *socket.Channel, arch Serve) string {
			result := launchScript(arch.Code)
			return result
		})

		if err != nil {
			log.Fatal(err)
		}
	}
}
