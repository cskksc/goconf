package main

import (
	"encoding/base64"
	"strings"
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (k KV) StartsWith(prefix string) bool {
	if strings.HasPrefix(k.Key, prefix) {
		return true
	}
	return false
}

// filter functions

type KVMatcher interface {
	Match(kv KV) bool
}

type StartsWithMatcher struct {
	prefix string
}

func (m StartsWithMatcher) Match(kv KV) bool {
	return kv.StartsWith(m.prefix)
}

type ExactMatcher struct {
	prefix string
}

func (m ExactMatcher) Match(kv KV) bool {
	return kv.Key == m.prefix
}

type DoesNotStartWithMatcher struct {
	prefixes []string
}

func (m DoesNotStartWithMatcher) Match(kv KV) bool {
	for _, p := range m.prefixes {
		if p != "" && kv.StartsWith(p) {
			return false
		}
	}
	return true
}

func filterKV(kvs <-chan KV, matcher KVMatcher) <-chan KV {
	out := make(chan KV, 10)
	go func() {
		for kv := range kvs {
			if ok := matcher.Match(kv); ok {
				out <- kv
			}
		}
		close(out)
	}()
	return out
}

// mapper functions

func base64ToStringKV(kv KV) KV {
	decKey, _ := base64.StdEncoding.DecodeString(kv.Key)
	decVal, _ := base64.StdEncoding.DecodeString(kv.Value)
	return KV{Key: string(decKey), Value: string(decVal)}
}

func stringKVToBase64(kv KV) KV {
	encKey := base64.StdEncoding.EncodeToString([]byte(kv.Key))
	encVal := base64.StdEncoding.EncodeToString([]byte(kv.Value))
	return KV{Key: string(encKey), Value: string(encVal)}
}

type mapperKVFunc func(kv KV) KV

func mapKV(kvs <-chan KV, fn mapperKVFunc) <-chan KV {
	out := make(chan KV, 10)
	go func() {
		for kv := range kvs {
			out <- fn(kv)
		}
		close(out)
	}()
	return out
}
