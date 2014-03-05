package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

var (
	errDoThis = errors.New("Error of doThis.")
	errDoThat = errors.New("Error of doThat.")
)

func init() {
    http.HandleFunc("/", errorHandler(betterHandler))
}

func errorHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        err := f(w, r)
        switch err {
        case errDoThis :
            http.Error(w, err.Error(), http.StatusForbidden)
        case errDoThat:
            http.Error(w, err.Error(), http.StatusNotFound)
        case nil:
        	fmt.Fprintf(w, "Hello World")
        default:
            http.Error(w, err.Error(), http.StatusInternalServerError)
            
        }

        log.Printf("handling %q: %v", r.RequestURI, err)
    }
}

func betterHandler(w http.ResponseWriter, r *http.Request) error {
    if err := doThis(); err != nil {
        return err
    }

    if err := doThat(); err != nil {
        return err
    }
    return nil
}

func doThis() error {
    if rand.Intn(100) > 50 {
        return errDoThis
    } else {
        return nil
    }
}

func doThat() error {
    if rand.Intn(100) > 50 {
        return errDoThat
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