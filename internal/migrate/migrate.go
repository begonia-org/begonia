package migrate

import (
	"fmt"

	app "github.com/begonia-org/go-sdk/api/app/v1"
	endpoint "github.com/begonia-org/go-sdk/api/endpoint/v1"
	file "github.com/begonia-org/go-sdk/api/file/v1"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/google/wire"

	"github.com/spark-lence/tiga"
)

var ProviderSet = wire.NewSet(NewMySQLMigrate,
	NewUsersOperator,
	NewTableModels,
	NewInitOperator,
	NewAPPOperator)

type TableModel interface{}
type MySQLMigrate struct {
	mysql  *tiga.MySQLDao
	models []TableModel
}

func NewTableModels() []TableModel {
	tables := make([]TableModel, 0)
	tables = append(tables, api.Users{}, endpoint.Endpoints{}, app.Apps{}, file.Files{}, file.Buckets{})
	return tables
}
func NewMySQLMigrate(mysql *tiga.MySQLDao, models ...TableModel) *MySQLMigrate {
	mysql.RegisterTimeSerializer()
	return &MySQLMigrate{mysql: mysql, models: models}
}

func (m *MySQLMigrate) Do() error {
	for _, model := range m.models {
		err := m.mysql.AutoMigrate(model)
		if err != nil {
			return fmt.Errorf("failed to migrate database: %w", err)
		}
	}
	return nil
}
