package pkg

import (
	"context"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/middleware"
	"github.com/begonia-org/begonia/internal/pkg/middleware/auth"
	"github.com/begonia-org/begonia/internal/pkg/migrate"
	"github.com/begonia-org/begonia/transport"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	config.NewConfig,
	// glc.NewKeyValueCache,
	// config.GetBlacklistPubSubChannel,
	// config.GetBlacklistPubSubGroup,
	// lbf.NewLayeredBloomFilter,
	context.Background,
	crypto.NewUsersAuth,
	migrate.NewMySQLMigrate,
	migrate.NewUsersOperator,
	migrate.NewTableModels,
	migrate.NewInitOperator,
	migrate.NewAPPOperator,
	auth.NewAccessKeyAuth,

	transport.NewLoggerMiddleware,
	middleware.New,
)
