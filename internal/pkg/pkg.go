package pkg

import (
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/crypto"
	"github.com/begonia-org/begonia/internal/pkg/middleware"
	"github.com/begonia-org/begonia/internal/pkg/migrate"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(config.NewConfig, middleware.NewAPIValidator, crypto.NewUsersAuth, migrate.NewMySQLMigrate, migrate.NewUsersOperator, migrate.NewTableModels, migrate.NewInitOperator)
