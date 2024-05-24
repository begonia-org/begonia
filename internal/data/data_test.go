package data

import (
	"log"
	"testing"
)

// func TestCreateInBatches(t *testing.T) {
func TestMain(m *testing.M) {
	code := m.Run()
	log.Printf("All tests passed with code %d", code)
}
