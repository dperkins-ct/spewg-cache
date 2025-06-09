package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
)

var port string
var peers string

func main() {
	flag.StringVar(&port, "port", ":8080", "HTTP Server Port")
	flag.StringVar(&peers, "peers", "", "comma-separated list of addresses")
	flag.Parse()

	peerList := strings.Split(peers, ",")

	cs := NewCacheServer(peerList)
	http.HandleFunc("/set", cs.SetHandler)
	http.HandleFunc("/get", cs.GetHandler)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
