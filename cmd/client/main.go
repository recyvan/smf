package main

import "flag"

func test_main() {
	serverconn := NewConn()
	serverconn.Connect("127.0.0.1:8080", "1234", "1234")
	Run(serverconn)
}
func main() {
	addr := flag.String("h", "127.0.0.1:8080", "server address")
	username := flag.String("u", "1234", "username")
	password := flag.String("p", "1234", "password")
	flag.Parse()
	client := NewConn()
	client.Connect(*addr, *username, *password)
	Run(client)
	//test_main()

}
