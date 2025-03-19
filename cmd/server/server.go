package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/recyvan/gotsgzengine/internal/command"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
)

type Conn struct {
	sync.Mutex
	User    []string
	conn    []net.Conn
	ConnMap map[string]net.Conn
}

type jsonMessage struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type response struct {
	Status string `json:"status"`
	ID     string `json:"id,omitempty"`
}

func NewConn() *Conn {
	return &Conn{
		User:    make([]string, 0),
		conn:    make([]net.Conn, 0),
		ConnMap: make(map[string]net.Conn),
	}
}

func (c *Conn) ListenAndServe(addr string, certFile string, keyFile string) {
	//初始化引擎
	engine := enginInit()
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		fmt.Println("[!] Error loading certificates:", err)
		os.Exit(1)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", addr, config)
	if err != nil {
		fmt.Println("[!] Error starting server:", err)
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Println("[-]Server is listening on port 8080...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("[!] Error connection:", err)
			continue
		}
		go c.ConnUserRegister(conn, engine)
	}
}

func (c *Conn) ConnUserRegister(conn net.Conn, engine *command.LocalEngine) {
	var tempdata jsonMessage
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&tempdata); err != nil {
		fmt.Println("[!] Error decoding JSON:", err)
		conn.Close()
		return
	}

	if checkToken(tempdata.Username, tempdata.Token) {
		connID := fmt.Sprintf("%s-%d", tempdata.Username, len(c.conn))
		resp := response{Status: "ok", ID: connID}
		respData, _ := json.Marshal(resp)
		conn.Write(append(respData, '\n'))

		c.Lock()
		c.conn = append(c.conn, conn)
		c.ConnMap[connID] = conn
		c.Unlock()

		fmt.Printf("[-] New user %s connected with ID %s\n", tempdata.Username, connID)
	} else {
		resp := response{Status: "error"}
		respData, _ := json.Marshal(resp)
		conn.Write(append(respData, '\n'))
		conn.Close()
	}
	Run(engine, conn)
}

func checkToken(username, token string) bool {
	file, err := os.Open("./token.txt")
	//data, err := ioutil.ReadFile("token.txt")
	if err != nil {
		return false
	}
	//读取token.txt文件，并按行分割
	data, err := ioutil.ReadAll(file)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == username+":"+token {
			return true
		}
	}
	return false
}
