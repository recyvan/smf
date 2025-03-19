package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Conn struct {
	conn     []net.Conn
	addr     []string
	ConnAddr map[string]net.Conn
	activeID string
	mu       sync.Mutex //还是没有搞清楚io重定向在终端的影响，这里通过AI询问解决了打印和输入冲突的问题，但是服务端还没有解决但是没关系可以正常运行
}

type jsonMessage struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type response struct {
	Status string `json:"status"`
	ID     string `json:"id"`
}

func NewConn() *Conn {
	return &Conn{
		conn:     make([]net.Conn, 0),
		addr:     make([]string, 0),
		ConnAddr: make(map[string]net.Conn),
	}
}

func (conn *Conn) Connect(addr, username, token string) {
	fmt.Println("Connecting to", addr)
	config := &tls.Config{InsecureSkipVerify: true}
	UserConn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	message := jsonMessage{Username: username, Token: token}
	marshal, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	_, err = UserConn.Write(marshal)
	if err != nil {
		fmt.Println("Error writing to connection:", err)
		return
	}
	fmt.Printf("Message sent successfully\n")

	tempscan := bufio.NewReader(UserConn)
	fmt.Println("Waiting for response...")
	res, err := tempscan.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	var res1 response
	err = json.Unmarshal([]byte(res), &res1)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	if res1.Status == "ok" {
		fmt.Printf("Connection established with ID: %s\n", res1.ID)
		conn.conn = append(conn.conn, UserConn)
		conn.ConnAddr[res1.ID] = UserConn
		conn.activeID = res1.ID

		go conn.handleServerMessages(UserConn)

	} else {
		fmt.Println("Failed to establish connection")
		err := UserConn.Close()
		if err != nil {
			return
		}
	}

}

func (conn *Conn) handleServerMessages(UserConn net.Conn) {
	serverScanner := bufio.NewScanner(UserConn)
	for serverScanner.Scan() {
		serverData := serverScanner.Text()
		conn.mu.Lock()
		fmt.Printf("\r%s\n", serverData)
		fmt.Printf("@%s->", conn.activeID)
		os.Stdout.Sync()
		conn.mu.Unlock()
	}
	if err := serverScanner.Err(); err != nil {
		conn.mu.Lock()
		fmt.Printf("[!] Error reading from server: %v\n", err)
		fmt.Printf("@%s->", conn.activeID)
		os.Stdout.Sync()
		conn.mu.Unlock()
	}
}

func (conn *Conn) ListConn() {
	conn.mu.Lock()
	fmt.Println("Active connections:")
	for id := range conn.ConnAddr {
		fmt.Println(id)
	}
	conn.mu.Unlock()
}

func (conn *Conn) CloseConn(connID string) {
	conn.mu.Lock()
	if connection, exists := conn.ConnAddr[connID]; exists {
		connection.Close()
		delete(conn.ConnAddr, connID)
		fmt.Printf("Connection %s closed\n", connID)
	} else {
		fmt.Printf("Connection %s does not exist\n", connID)
	}
	conn.mu.Unlock()
}

func (conn *Conn) ChangeConn(connID string) {
	conn.mu.Lock()
	if _, exists := conn.ConnAddr[connID]; exists {
		conn.activeID = connID
		fmt.Printf("Switched to connection %s\n", connID)
	} else {
		fmt.Printf("Connection %s does not exist\n", connID)
	}
	conn.mu.Unlock()
}

func Run(conn *Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Fprintf(os.Stdout, "The management commands for conn connection are: listconn, closeconn, changeconn!\n")
	fmt.Fprintf(os.Stdout, "@%s->", conn.activeID)
	os.Stdout.Sync()
	for scanner.Scan() {
		input := scanner.Text()
		parts := strings.Fields(input)
		if len(parts) == 0 {
			fmt.Fprintf(os.Stdout, "@%s->", conn.activeID)
			os.Stdout.Sync()
			continue
		}

		switch parts[0] {
		case "listconn":
			conn.ListConn()
		case "closeconn":
			if len(parts) < 2 {
				fmt.Printf("@%s-> Usage: closeconn <conn.ID>\n", conn.activeID)
				fmt.Fprintf(os.Stdout, "@%s->", conn.activeID)
				os.Stdout.Sync()
				continue
			}
			conn.CloseConn(parts[1])
		case "changeconn":
			if len(parts) < 2 {
				fmt.Printf("@%s-> Usage: changeconn <conn.ID>\n", conn.activeID)
				fmt.Fprintf(os.Stdout, "@%s->", conn.activeID)
				os.Stdout.Sync()
				continue
			}
			conn.ChangeConn(parts[1])
		default:
			if UserConn, exists := conn.ConnAddr[conn.activeID]; exists {
				if _, err := fmt.Fprintln(UserConn, input); err != nil {
					fmt.Printf("[!] Error sending to server: %v\n", err)
					fmt.Fprintf(os.Stdout, "@%s->", conn.activeID)
					os.Stdout.Sync()
					break
				}
				if input == "exit" {
					break
				}
			} else {
				fmt.Printf("No active connection\n")
			}
		}
		fmt.Fprintf(os.Stdout, "@%s->", conn.activeID)
		os.Stdout.Sync()
	}
}
