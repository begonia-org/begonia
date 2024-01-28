package pkg

import (
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/middleware"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(config.NewConfig, middleware.NewAPIVildator, crypto.NewUsersAuth)
