package pkg

import (
	"github.com/google/wire"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"github.com/wetrycode/begonia/internal/pkg/crypto"
	"github.com/wetrycode/begonia/internal/pkg/middleware"
)

var ProviderSet = wire.NewSet(config.NewConfig, middleware.NewAPIVildator, crypto.NewUsersAuth)
