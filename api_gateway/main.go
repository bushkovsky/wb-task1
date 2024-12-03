package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	flushTimeout = 7000
)

var topic1 = "orders"
var topic2 = "get-by-id"
var topic3 = "get-order"

func main() {

	prod, err := NewProd()
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %s", err)
	}
	defer prod.Close()

	cons, err := NewConsumer(topic3)
	chanOrder := make(chan Order)

	http.HandleFunc("/get-order-id", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var id OrderID
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&id); err != nil {
			http.Error(w, "Failed to decode JSON body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if id.OrderUID == "" {
			http.Error(w, "Missing required field: order_uid", http.StatusBadRequest)
			return
		}

		message, err := json.Marshal(id)
		if err != nil {
			http.Error(w, "Failed to encode order to JSON", http.StatusInternalServerError)
			return
		}

		err = prod.Produce(message, topic2)

		go func() {
			cons.Start(chanOrder)
		}()

		select {
		case <-time.After(13 * time.Second):
			log.Printf("too long time")
		case ord := <-chanOrder:
			res, err := json.Marshal(ord)
			if err != nil {
				http.Error(w, "Failed to encode order to JSON after getting", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(http.StatusOK)

			_, err = w.Write(res)
			if err != nil {
				http.Error(w, "Failed to write order to JSON after getting", http.StatusBadRequest)
				return
			}

		}

		if err != nil {
			http.Error(w, "Failed to produce message", http.StatusBadRequest)
			log.Fatalf("Failed to produce message: %v", err)
		}

	})

	http.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var order Order
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&order); err != nil {
			http.Error(w, "Failed to decode JSON body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if order.OrderUID == "" {
			http.Error(w, "Missing required field: order_uid", http.StatusBadRequest)
			return
		}

		message, err := json.Marshal(order)
		if err != nil {
			http.Error(w, "Failed to encode order to JSON", http.StatusInternalServerError)
			return
		}

		err = prod.Produce(message, topic1)
		if err != nil {
			http.Error(w, "Failed to produce message", http.StatusInternalServerError)
			log.Fatalf("Failed to produce message: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		log.Println(w, `{"status":"Order submitted successfully"}`)
	})

	fmt.Println("API Gateway running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
