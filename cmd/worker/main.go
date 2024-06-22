package main

import (
	"crypto/rand"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"io"
	"log"
	"net/http"
	"time"
)

var decoder = schema.NewDecoder()
var staticData = make([]byte, MaxChunkSize)

const (
	MaxChunkSize = 16 << 20 // 16 MB
	DefChunkSize = 1 << 20  // 1 MB
	DefCount     = 10
)

type Stream1Data struct {
	Chunk int `json:"chunk"`
	Count int `json:"count"`
}

func init() {
	_, err := rand.Read(staticData)
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/stream1", stream1)
	http.HandleFunc("/sleep", sleep)
	http.ListenAndServe("0.0.0.0:8080", logRequest(http.DefaultServeMux))
}

func stream1(w http.ResponseWriter, req *http.Request) {
	data := Stream1Data{Chunk: DefChunkSize, Count: DefCount}
	err := decoder.Decode(&data, req.URL.Query())
	if data.Chunk > MaxChunkSize {
		data.Chunk = MaxChunkSize
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(data)
	for i := 1; i <= data.Count; i++ {
		if _, err := w.Write([]byte("\u0000")); err != nil {
			if err == io.EOF {
				break
			}
			w.Write([]byte(err.Error()))
			return
		}
		if _, err := w.Write(staticData[:data.Chunk]); err != nil {
			if err == io.EOF {
				break
			}
			w.Write([]byte(err.Error()))
			return
		}
	}
	w.Write([]byte("OK"))
}

func sleep(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("sleeping for 1 minute"))
	time.Sleep(60 * time.Second)
	w.Write([]byte("OK"))
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New()
		t1 := time.Now()
		log.Printf("IN %v %s %s %s %s\n", id, r.RemoteAddr, r.Method, r.URL, r.UserAgent())
		handler.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("OUT %v %s %s %s %v\n", id, r.RemoteAddr, r.Method, r.URL, t2.Sub(t1))
	})
}
