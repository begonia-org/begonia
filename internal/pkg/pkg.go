package pkg

import (
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/middleware/validator"
	"github.com/begonia-org/begonia/internal/pkg/migrate"
	lbf "github.com/begonia-org/go-layered-bloom"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	config.NewConfig,
	lbf.NewBloomPubSub,
	config.GetBlacklistPubSubChannel,
	config.GetBlacklistPubSubGroup,
	lbf.NewLayeredBloomFilter,
	validator.NewAPIValidator,
	crypto.NewUsersAuth,
	migrate.NewMySQLMigrate,
	migrate.NewUsersOperator,
	migrate.NewTableModels,
	migrate.NewInitOperator,
)
