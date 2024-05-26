package pkg

import (
	"context"

	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	config.NewConfig,
	context.Background,
	crypto.NewUsersAuth,

	gateway.NewLoggerMiddleware,
)
