package migrate

import (
	"log"

	"github.com/begonia-org/begonia/internal/pkg/config"
)

type InitOperator struct {
	migrate *MySQLMigrate
	user    *UsersOperator
	app    *APPOperator
	config  *config.Config
}

func NewInitOperator(migrate *MySQLMigrate, user *UsersOperator,app *APPOperator, config *config.Config) *InitOperator {
	return &InitOperator{migrate: migrate, user: user, config: config,app:app}
}

func (m *InitOperator) Init() error {
	err := m.migrate.Do()
	if err != nil {
		log.Printf("failed to migrate database: %v", err)
		return err
	}
	adminPasswd := m.config.GetDefaultAdminPasswd()
	name := m.config.GetDefaultAdminName()
	email := m.config.GetDefaultAdminEmail()
	phone := m.config.GetDefaultAdminPhone()
	aseKey := m.config.GetAesKey()
	ivKey := m.config.GetAesIv()
	uid,err := m.user.InitAdminUser(adminPasswd, aseKey, ivKey, name, email, phone)
	if err != nil {
		log.Printf("failed to init admin user: %v", err)
		return err
	}

	return m.app.InitAdminAPP(uid)
}
