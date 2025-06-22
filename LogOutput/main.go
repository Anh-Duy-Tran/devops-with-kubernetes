package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

func main() {
	id := uuid.New()
	for {
		currentTime := time.Now()
		fmt.Printf("[%s]: %s\n", currentTime.Format(time.RFC3339), id)
		time.Sleep(5 * time.Second)
	}
}
