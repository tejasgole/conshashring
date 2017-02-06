// Consistent Hash Ring implementation
// HTTP service to create/query CHR
// Copyright Jan 2017
// Author: Abhijeet Gole

package main

import (
	"fmt"
	"bptree"
	"crypto/md5"
	"encoding/binary"
	"bytes"
	"regexp"
	"net/http"
	"strconv"
)

var validPath = regexp.MustCompile("^/(add|get|del|getN|/)/([a-zA-Z0-9.]+)$")
var btree *bptree.Bptree

func createHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	btree, err = bptree.New(3)
	if err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}
	fmt.Fprintf(w, "Created Ring\n")
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		fmt.Fprintf(w, "Invalid\n")
		return
	}
	if btree == nil {
		fmt.Fprintf(w, "Ring not created\n")
		return
	}
	//cq := crc64.MakeTable(0xD5828281)
	//fmt.Println(m[2])
	//key := crc64.Checksum([]byte(m[2]), cq)
	kmd5 := md5.Sum([]byte(m[2]))
	var key uint64
	_ = binary.Read(bytes.NewReader(kmd5[0:8]), binary.BigEndian, &key)
	btree.Insert(bptree.Item(key), m[2])
	fmt.Fprintf(w, "Added %s, key:%x\n", m[2], key)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		fmt.Fprintf(w, "Invalid\n")
		return
	}
	if btree == nil {
		fmt.Fprintf(w, "Ring not created\n")
		return
	}
	//fmt.Println(m[2])
	//cq := crc64.MakeTable(0xD5828281)
	//key := crc64.Checksum([]byte(m[2]), cq)
	kmd5 := md5.Sum([]byte(m[2]))
	var key uint64
	_ = binary.Read(bytes.NewReader(kmd5[0:8]), binary.BigEndian, &key)
	val := btree.Get(bptree.Item(key))
	fmt.Fprintf(w, "key:%x,val:%s\n", key, val)
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		fmt.Fprintf(w, "Invalid\n")
		return
	}
	if btree == nil {
		fmt.Fprintf(w, "Ring not created\n")
		return
	}
	//fmt.Println(m[2])
	//cq := crc64.MakeTable(0xD5828281)
	//key := crc64.Checksum([]byte(m[2]), cq)
	kmd5 := md5.Sum([]byte(m[2]))
	var key uint64
	_ = binary.Read(bytes.NewReader(kmd5[0:8]), binary.BigEndian, &key)
	_, val := btree.Del(bptree.Item(key))
	fmt.Fprintf(w, "key:%x,val:%s\n", key, val)
}

func printHandler(w http.ResponseWriter, r *http.Request) {
	if btree == nil {
		fmt.Fprintf(w, "Ring not created\n")
		return
	}
	//fmt.Println(m[2])
	btree.Print(w)
}

func getNHandler(w http.ResponseWriter, r *http.Request) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		fmt.Fprintf(w, "Invalid\n")
		return
	}
	if btree == nil {
		fmt.Fprintf(w, "Ring not created\n")
		return
	}
	//fmt.Println(m[2])
	//cq := crc64.MakeTable(0xD5828281)
	key, err := strconv.Atoi(m[2])
	if err != nil {
		fmt.Fprintf(w, "Invalid key\n")
		return
	}
	val := btree.GetNextN(bptree.Item(key), 3)
	fmt.Fprintf(w, "key:%d,val:%s\n", key, val)
}

func main() {

	http.HandleFunc("/creat", createHandler)
	http.HandleFunc("/add/", addHandler)
	http.HandleFunc("/get/", getHandler)
	http.HandleFunc("/del/", delHandler)
	http.HandleFunc("/getN/", getNHandler)
	http.HandleFunc("/print", printHandler)
	http.ListenAndServe(":8080", nil)
}
