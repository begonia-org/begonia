package main

import (
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"github.com/begonia-org/begonia/internal/server"
)

func main() {
	config := config.ReadConfig("dev")
	server := server.New(config, logger.Logger, "0.0.0.0:12138")
	// go func() {
	// 	err := server.Start()
	// 	t.Error(err)
	// }()
	server.Start()

}
