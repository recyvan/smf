package main

import "flag"

func main() {
	host := flag.String("h", "127.0.0.1", "host ip address")
	port := flag.String("p", "8080", "host port")
	username := flag.String("u", "1234", "username")
	password := flag.String("p", "1234", "password")
	flag.Parse()
	client := NewConn()
	server := *host + ":" + *port
	client.Connect(server, *username, *password)
	Run(client)

}
