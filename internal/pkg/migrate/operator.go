package migrate

import "github.com/begonia-org/begonia/internal/pkg/config"

type InitOperator struct {
	migrate *MySQLMigrate
	user    *UsersOperator
	config  *config.Config
}

func NewInitOperator(migrate *MySQLMigrate, user *UsersOperator, config *config.Config) *InitOperator {
	return &InitOperator{migrate: migrate, user: user, config: config}
}

func (m *InitOperator) Init() error {
	err := m.migrate.Do()
	if err != nil {
		return err
	}
	adminPasswd := m.config.GetDefaultAdminPasswd()
	name := m.config.GetDefaultAdminName()
	email := m.config.GetDefaultAdminEmail()
	phone := m.config.GetDefaultAdminPhone()
	aseKey := m.config.GetAesKey()
	ivKey := m.config.GetAesIv()
	err = m.user.InitAdminUser(adminPasswd, aseKey, ivKey, name, email, phone)
	return err
}
