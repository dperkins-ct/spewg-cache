package spewg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const replicationHeader = "X-Replication"

type CacheServer struct {
	cache    *Cache
	peers    []string
	hashRing *HashRing
	selfID   string
	mu       sync.Mutex
}

func NewCacheServer(peers []string, selfID string) *CacheServer {
	cache := NewCache(10)
	cache.startEvictionTicker(1 * time.Minute)
	cs := &CacheServer{
		cache:    cache,
		peers:    peers,
		hashRing: NewHashRing(),
		selfID:   selfID,
	}
	for _, peer := range peers {
		cs.hashRing.AddNode(Node{ID: peer, Addr: peer})
	}
	cs.hashRing.AddNode(Node{ID: selfID, Addr: "self"})
	return cs
}

func (cs *CacheServer) SetHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	targetNode := cs.hashRing.GetNode(req.Key)
	if targetNode.Addr == "self" {
		cs.cache.Set(req.Key, req.Value, 1*time.Hour)
		if r.Header.Get(replicationHeader) == "" {
			go cs.replicateSet(req.Key, req.Value)
		}
		w.WriteHeader(http.StatusOK)
	} else {
		cs.forwardRequest(w, targetNode, r)
	}
}

func (cs *CacheServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	targetNode := cs.hashRing.GetNode(key)

	if targetNode.Addr == "self" {
		value, found := cs.cache.Get(key)
		if !found {
			http.NotFound(w, r)
			return
		}
		err := json.NewEncoder(w).Encode((map[string]string{
			"value": value,
		}))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		originalSender := r.Header.Get("X-Forwarded-For")
		if originalSender == cs.selfID {
			http.Error(w, "Loop Detected", http.StatusBadRequest)
			return
		}
		r.Header.Set("X-Forwarded-For", cs.selfID)
		cs.forwardRequest(w, targetNode, r)
	}
}

func (cs *CacheServer) replicateSet(key, value string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	req := struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		Key:   key,
		Value: value,
	}

	data, _ := json.Marshal(req)

	for _, peer := range cs.peers {
		if peer != cs.selfID {
			go func(peer string) {
				client := &http.Client{}
				req, err := http.NewRequest("POST", peer+"/set", bytes.NewReader(data))
				if err != nil {
					log.Printf("failed to created replication request: %v", err)
					return
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set(replicationHeader, "true")
				_, err = client.Do(req)
				if err != nil {
					log.Printf("Failed to replicate to peer %s: %v", peer, err)
				}
				log.Println("replication successful to", peer)
			}(peer)
		}
	}
}

func (cs *CacheServer) forwardRequest(w http.ResponseWriter, targetNode Node, r *http.Request) {
	client := &http.Client{}

	var req *http.Request
	var err error

	if r.Method == http.MethodGet {
		url := fmt.Sprintf("%s%s?%s", targetNode.Addr, r.URL.Path, r.URL.RawQuery)
		req, err = http.NewRequest(r.Method, url, nil)
	} else if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		url := fmt.Sprintf("%s%s", targetNode.Addr, r.URL.Path)
		req, err = http.NewRequest(r.Method, url, bytes.NewReader(body))
		if err != nil {
			http.Error(w, "failed to create forward request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
	}

	if err != nil {
		log.Printf("Failed to create forward request: %v", err)
		http.Error(w, "Failed to create forward request", http.StatusInternalServerError)
		return
	}

	req.Header = r.Header

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to forward request to node %s: %v", targetNode.Addr, err)
		http.Error(w, "Failed to forward request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Failed to write response from node %s: %v", targetNode.Addr, err)
	}
}
