//func main() {
//	listener, err := net.Listen("tcp", ":8080")
//	if err != nil {
//		fmt.Println("Error starting localserver:", err)
//		os.Exit(1)
//	}
//	defer listener.Close()
//
//	fmt.Println("Server is listening on port 8080...")
//
//	for {
//		conn, err := listener.Accept()
//		if err != nil {
//			fmt.Println("Error accepting connection:", err)
//			continue
//		}
//
//		go Run(conn)
//	}
//}

package main

func test_main() {
	serverconn := NewConn()
	serverconn.ListenAndServe(":8080", "server.crt", "server.key")

}
func main() {
	//certFile := flag.String("sc", "server.crt", "Path to the server certificate")
	//keyFile := flag.String("sk", "server.key", "Path to the server key")
	//port := flag.Int("p", 8080, "Port to listen on")
	//flag.Parse()
	//serverconn := NewConn()
	//server_host := "0.0.0.0" + ":" + strconv.Itoa(*port)
	//serverconn.ListenAndServe(server_host, *certFile, *keyFile)
	test_main()
}
