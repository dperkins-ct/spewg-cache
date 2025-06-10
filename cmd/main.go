package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"spewg-cache"
	"strings"
)

var port string
var peers string

func main() {
	flag.StringVar(&port, "port", ":8080", "HTTP Server Port")
	flag.StringVar(&peers, "peers", "", "comma-separated list of addresses")
	flag.Parse()

	nodeID := fmt.Sprintf("%s%d", "node", rand.Intn(100))
	peerList := strings.Split(peers, ",")

	cs := spewg.NewCacheServer(peerList, nodeID)
	http.HandleFunc("/set", cs.SetHandler)
	http.HandleFunc("/get", cs.GetHandler)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
