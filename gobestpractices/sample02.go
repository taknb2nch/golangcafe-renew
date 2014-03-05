package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

func init() {
    http.HandleFunc("/", errorHandler(betterHandler))
}

func errorHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        err := f(w, r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            log.Printf("handling %q: %v", r.RequestURI, err)
        }
    }
}

func betterHandler(w http.ResponseWriter, r *http.Request) error {
    if err := doThis(); err != nil {
        return fmt.Errorf("doing this: %v", err)
    }

    if err := doThat(); err != nil {
        return fmt.Errorf("doing that: %v", err)
    }
    return nil
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