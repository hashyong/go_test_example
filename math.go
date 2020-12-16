package main

import "fmt"

func MyMod(a int, b int) int {
	if b == 0 {
		return -1
	}

	return a % b
}

// db.go
type DB interface {
	Get(key string) (int, error)
}

func GetFromDB(db DB, key string) int {
	if value, err := db.Get(key); err == nil {
		return value
	}

	return -1
}

func main() {
	fmt.Print("hello")
}
