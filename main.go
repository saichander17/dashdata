package main

import (
// 	"fmt",
	"time"
)

// func main() {
// 	store := NewStore()
//
// 	// Set a value
// 	store.Set("name", "Dashdata")
//
// 	// Get a value
// 	if value, exists := store.Get("name"); exists {
// 		fmt.Printf("Value for key 'name': %s\n", value)
// 	} else {
// 		fmt.Println("Key 'name' not found")
// 	}
//
// 	// Update a value
// 	store.Set("name", "Dashdata v1.0")
//
// 	// Get the updated value
// 	if value, exists := store.Get("name"); exists {
// 		fmt.Printf("Updated value for key 'name': %s\n", value)
// 	}
//
// 	// Delete a value
// 	store.Delete("name")
//
// 	// Try to get the deleted value
// 	if _, exists := store.Get("name"); !exists {
// 		fmt.Println("Key 'name' was successfully deleted")
// 	}
// }

func main() {
	store := NewStore()
	server := NewServer(store, "8080", 1000, 10000, 5*time.Second)
	err := server.Start()
	if err != nil {
		panic(err)
	}
}
