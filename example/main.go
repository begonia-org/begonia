package main

import "github.com/begonia-org/go-sdk/example"

func main() {
	go example.Run("0.0.0.0:1949")
	example.Run("0.0.0.0:2024")

}
