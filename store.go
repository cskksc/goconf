package main

import "io"

type Store interface {
	GetAllKV() <-chan KV
	GetAllACL() <-chan KV

	GetACL(token string) []string

	SetKV(kv KV) error
	SetACL(kv KV) error

	DeleteKV(kv KV) error
	DeleteACL(kv KV) error

	Backup(w io.Writer) (int, error)
}
