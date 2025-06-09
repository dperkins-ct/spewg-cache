package spewg

import (
	"bytes"
	"encoding/json"
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

func NewCacheServer(peers []string) *CacheServer {
	cache := NewCache(10)
	cache.startEvictionTicker(1 * time.Minute)
	return &CacheServer{
		cache: cache,
		peers: peers,
	}
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
	// Set cache entry with a default expiration time, 1 hour
	cs.cache.Set(req.Key, req.Value, 1*time.Hour)

	if r.Header.Get(replicationHeader) == "" {
		go cs.replicateSet(req.Key, req.Value)
	}
	w.WriteHeader(http.StatusOK)
}

func (cs *CacheServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value, found := cs.cache.Get(key)
	if !found {
		http.NotFound(w, r)
		return
	}
	json.NewEncoder(w).Encode((map[string]string{
		"value": value,
	}))
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
