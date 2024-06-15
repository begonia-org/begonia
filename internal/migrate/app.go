package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type APPOperator struct {
	mysql *tiga.MySQLDao
}

func NewAPPOperator(mysql *tiga.MySQLDao) *APPOperator {
	return &APPOperator{mysql: mysql}
}
func dumpInitApp(app *api.Apps) error {
	log.Print("########################################admin-app###############################")
	log.Printf("Init appid:%s", app.Appid)
	log.Printf("Init accessKey:%s", app.AccessKey)
	log.Printf("Init secret:%s", app.Secret)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(homeDir, ".begonia")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(path, "admin-app.json"))
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(app); err != nil {
		return err
	}
	log.Printf("Init admin-app config file at :%s", file.Name())
	log.Print("#################################################################################")
	return nil
}
func (m *APPOperator) InitAdminAPP(owner string) (err error) {
	app := &api.Apps{}
	defer func() {
		if app.Appid != "" {
			if errInit := dumpInitApp(app); errInit != nil {
				err = errInit
			}
		}
	}()
	err = m.mysql.First(context.TODO(), app, "name = ?", "admin-app")
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if app.Appid == "" {
		snk, errSnk := tiga.NewSnowflake(1)
		if errSnk != nil {
			err = errSnk
			return err
		}
		accessKey, errAccess := biz.GenerateAppAccessKey()

		secret, errSecret := biz.GenerateAppSecret()

		if errAccess != nil || errSecret != nil {
			err = fmt.Errorf("failed to generate accessKey or secret: %v,%v", errAccess, errSecret)
			return err
		}

		appid := biz.GenerateAppid(snk)

		app = &api.Apps{
			Appid:       appid,
			AccessKey:   accessKey,
			Secret:      secret,
			Name:        "admin-app",
			Description: "admin app",
			Owner:       owner,
			Status:      api.APPStatus_APP_ENABLED,
			CreatedAt:   timestamppb.New(time.Now()),
			UpdatedAt:   timestamppb.New(time.Now()),
			Tags:        []string{"admin"},
		}
		err = m.mysql.Create(context.Background(), app)
		return err
	}
	return nil
}
