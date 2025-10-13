package main

import (
	"log"

	"GateKeeper/configurations"
)

func main() {
	log.Println("Testing Supabase connection...")

	// Connect to Supabase
	configurations.ConnectToSupabase()

	log.Println("🎉 Connection test completed successfully!")
}
