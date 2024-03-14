package pkg

import (
	"context"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/middleware"
	"github.com/begonia-org/begonia/internal/pkg/middleware/validator"
	"github.com/begonia-org/begonia/internal/pkg/migrate"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	config.NewConfig,
	// glc.NewKeyValueCache,
	// config.GetBlacklistPubSubChannel,
	// config.GetBlacklistPubSubGroup,
	// lbf.NewLayeredBloomFilter,
	context.Background,
	validator.NewAPIValidator,
	crypto.NewUsersAuth,
	migrate.NewMySQLMigrate,
	migrate.NewUsersOperator,
	migrate.NewTableModels,
	migrate.NewInitOperator,

	middleware.NewLoggerMiddleware,
)
