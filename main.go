package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
	"io/fs"

	"github.com/boltdb/bolt"
)
//go:embed web/*
var webFS embed.FS

type StatusEntry struct {
	Timestamp int64  `json:"timestamp"`
	Up        bool   `json:"up"`
}

type HostStatus struct {
	Hostname string        `json:"hostname"`
	History  []StatusEntry `json:"history"`
}

var db *bolt.DB

const (
	bucketName = "status"
	maxDays    = 14
)

func main() {
	var err error
	db, err = bolt.Open("status.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/api/ping", handlePing)
	http.HandleFunc("/api/hosts", handleHosts)
	http.HandleFunc("/api/history", handleHistory)

	// Serve files from embedded web/ directory at root, hiding /web/
       fs, err := fsSub(webFS, "web")
       if err != nil {
	       log.Fatal(err)
       }
       http.Handle("/", http.FileServer(http.FS(fs)))

       fmt.Println("Server running on :8086 (dynamic subpath, embedded web directory)")
       log.Fatal(http.ListenAndServe(":8086", nil))

// End of main
}

// Helper for Go <1.16.7 compatibility
func fsSub(fsys embed.FS, dir string) (fs.FS, error) {
	return fs.Sub(fsys, dir)
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
       var req struct {
	       Hostname string `json:"hostname"`
       }
       if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	       w.WriteHeader(http.StatusBadRequest)
	       return
       }
       if req.Hostname == "" {
	       w.WriteHeader(http.StatusBadRequest)
	       return
       }
       storeStatus(req.Hostname, true)
       w.WriteHeader(http.StatusOK)
}

func storeStatus(hostname string, up bool) {
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		var history []StatusEntry
		v := b.Get([]byte(hostname))
		if v != nil {
			json.Unmarshal(v, &history)
		}
		entry := StatusEntry{Timestamp: time.Now().Unix(), Up: up}
		history = append(history, entry)
		cutoff := time.Now().AddDate(0, 0, -maxDays).Unix()
		var filtered []StatusEntry
		for _, e := range history {
			if e.Timestamp >= cutoff {
				filtered = append(filtered, e)
			}
		}
		data, _ := json.Marshal(filtered)
		return b.Put([]byte(hostname), data)
	})
}

func handleHosts(w http.ResponseWriter, r *http.Request) {
       type HostInfo struct {
	       Hostname   string `json:"hostname"`
	       LastSeen   int64  `json:"last_seen"`
       }
       var hosts []HostInfo
       db.View(func(tx *bolt.Tx) error {
	       b := tx.Bucket([]byte(bucketName))
	       if b == nil {
		       return nil
	       }
	       return b.ForEach(func(k, v []byte) error {
		       var history []StatusEntry
		       json.Unmarshal(v, &history)
		       var lastSeen int64
		       if len(history) > 0 {
			       lastSeen = history[len(history)-1].Timestamp
		       }
		       hosts = append(hosts, HostInfo{Hostname: string(k), LastSeen: lastSeen})
		       return nil
	       })
       })
       sort.Slice(hosts, func(i, j int) bool { return hosts[i].Hostname < hosts[j].Hostname })
       json.NewEncoder(w).Encode(hosts)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	hostname := r.URL.Query().Get("hostname")
	if hostname == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var history []StatusEntry
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(hostname))
		if v != nil {
			json.Unmarshal(v, &history)
		}
		return nil
	})
	json.NewEncoder(w).Encode(history)
}
