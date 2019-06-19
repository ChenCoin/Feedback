package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	db, err := bolt.Open("./my.db", 0600, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	_ = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		fmt.Println("CreateBucketIfNotExists bucket suc")
		return nil
	})
	db.Close()

	http.HandleFunc("/get/", getFeedback)
	http.HandleFunc("/put/", putFeedback)
	defer fmt.Println("server close")
	err = http.ListenAndServe(":8090", nil)
	if err != nil {
		fmt.Println("error when create server")
	}
}

type Message struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func putFeedback(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(404)
		_, _ = fmt.Fprintf(w, "")
		return
	}

	var msg = Message{}
	msg.Title = strings.Join(r.Form["title"], "")
	msg.Content = strings.Join(r.Form["content"], "")
	if len(msg.Title) > 64 || len(msg.Content) > 2048 {
		w.WriteHeader(404)
		_, _ = fmt.Fprintf(w, "")
		return
	}
	msgString, err := json.Marshal(msg)
	if err != nil {
		_ = fmt.Errorf("create bucket: %s", err)
	}
	fmt.Println("put ", msg.Title, msg.Content)

	db, err := bolt.Open("./my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		id, _ := b.NextSequence()
		fmt.Println("d ", msgString)
		err := b.Put(i2b(int(id)), []byte(msgString))
		_, _ = fmt.Fprintf(w, "{\"statue\":\"success\"}")
		return err
	})
}

func getFeedback(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(404)
		_, _ = fmt.Fprintf(w, "")
		return
	}
	pageString := strings.Join(r.Form["page"], "")
	page, err := strconv.Atoi(pageString)
	if err != nil || page < 1 {
		w.WriteHeader(404)
		_, _ = fmt.Fprintf(w, "")
		return
	}
	fmt.Println("get")

	db, err := bolt.Open("./my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		c := b.Cursor()
		index := 0
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if index >= page*10 {
				break
			}
			if index == (page-1)*10 {
				buffer.Write(v)
			}
			if index > (page-1)*10 {
				buffer.WriteString(",")
				buffer.Write(v)
			}
			index++
		}
		return nil
	})
	buffer.WriteString("]")

	if err != nil {
		w.WriteHeader(404)
		_, _ = fmt.Fprintf(w, "")
		return
	}
	_, _ = fmt.Fprintf(w, buffer.String())
}

func i2b(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
