package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

func init() {
    http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
    err := doThis()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Printf("handling %q: %v", r.RequestURI, err)
        return
    }

    err = doThat()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Printf("handling %q: %v", r.RequestURI, err)
        return
    }
}

func doThis() error {
    if rand.Intn(100) > 50 {
        return fmt.Errorf("Error of doThis.")
    } else {
        return nil
    }
}

func doThat() error {
    if rand.Intn(100) > 50 {
        return fmt.Errorf("Error of doThat.")
    } else {
        return nil
    }
}

func main() {
    err := http.ListenAndServe(":9090", nil)
    if err != nil {
        log.Fatalf("%v\n", err)
    }
}