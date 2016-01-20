package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func (s *Server) serveGetKV(w http.ResponseWriter, r *http.Request) {
	// create a filter to remove kv's protected by ACL
	token := r.URL.Query().Get("token")
	acls := s.store.GetACL(token)
	aclFilter := DoesNotStartWithMatcher{prefixes: acls}

	// create a matcher depending on recurse option
	recurse := r.URL.Query().Get("recurse")
	prefix := kvPattern.ReplaceAllString(r.URL.Path, "$2")
	var matcher KVMatcher
	if recurse != "" {
		matcher = StartsWithMatcher{prefix: prefix}
	} else {
		matcher = ExactMatcher{prefix: prefix}
	}

	kvs := make([]KV, 0)
	for kv := range filterKV(filterKV(s.store.GetAllKV(), matcher), aclFilter) {
		kvs = append(kvs, kv)
	}
	j, _ := json.Marshal(kvs)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(j))
}

func (s *Server) servePostKV(w http.ResponseWriter, r *http.Request) {
	// create a filter to remove kv's protected by ACL
	token := r.URL.Query().Get("token")
	acls := s.store.GetACL(token)
	aclFilter := DoesNotStartWithMatcher{prefixes: acls}

	var kvs []KV
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &kvs)

	kvChan := make(chan KV, 10)
	go func() {
		for _, kv := range kvs {
			kvChan <- kv
		}
		close(kvChan)
	}()

	for kv := range filterKV(kvChan, aclFilter) {
		if err := s.store.SetKV(kv); err != nil {
			log.Println("[ERR] " + err.Error())
		}
	}
}

func (s *Server) serveDeleteKV(w http.ResponseWriter, r *http.Request) {
	// create a filter to remove kv's protected by ACL
	token := r.URL.Query().Get("token")
	acls := s.store.GetACL(token)
	aclFilter := DoesNotStartWithMatcher{prefixes: acls}

	var keys []string
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &keys)

	kvChan := make(chan KV, 10)
	go func() {
		for _, k := range keys {
			kvChan <- KV{Key: k, Value: ""}
		}
		close(kvChan)
	}()

	for kv := range filterKV(kvChan, aclFilter) {
		if err := s.store.DeleteKV(kv); err != nil {
			log.Println("[ERR] " + err.Error())
		}
	}
}