package migrate

import (
	"fmt"

	api "github.com/begonia-org/go-sdk/api/user/v1"
	endpoint "github.com/begonia-org/go-sdk/api/endpoint/v1"
	app "github.com/begonia-org/go-sdk/api/app/v1"

	"github.com/spark-lence/tiga"
)

type TableModel interface{}
type MySQLMigrate struct {
	mysql  *tiga.MySQLDao
	models []TableModel
}

func NewTableModels() []TableModel {
	tables := make([]TableModel, 0)
	tables = append(tables, api.Users{}, endpoint.Endpoints{}, app.Apps{})
	return tables
}
func NewMySQLMigrate(mysql *tiga.MySQLDao, models ...TableModel) *MySQLMigrate {
	mysql.RegisterTimeSerializer()
	return &MySQLMigrate{mysql: mysql, models: models}
}
func (m *MySQLMigrate) BindModel(model interface{}) {
	m.models = append(m.models, model)
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
