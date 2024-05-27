package migrate

import (
	"context"
	"encoding/json"
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
func dumpInitApp(app *api.Apps) {
	log.Print("########################################admin-app###############################")
	log.Printf("Init appid:%s", app.Appid)
	log.Printf("Init accessKey:%s", app.AccessKey)
	log.Printf("Init secret:%s", app.Secret)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	path := filepath.Join(homeDir, ".begonia")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Fatalf(err.Error())
		return
	}
	file, err := os.Create(filepath.Join(path, "admin-app.json"))
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(app); err != nil {
		log.Fatalf(err.Error())
		return
	}
	log.Printf("Init admin-app config file at :%s", file.Name())
	log.Print("#################################################################################")

}
func (m *APPOperator) InitAdminAPP(owner string) error {
	app := &api.Apps{}
	defer func() {
		if app.Appid != "" {
			dumpInitApp(app)
		}
	}()
	err := m.mysql.First(context.TODO(), app, "name = ?", "admin-app")
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Fatalf("InitAdminAPP error:%v", err)
		return err
	}
	if app.Appid == "" {
		snk, err := tiga.NewSnowflake(1)
		if err != nil {
			return err
		}
		// 初始化数据
		// uid := snk.GenerateID()
		accessKey, err := biz.GenerateAppAccessKey()
		if err != nil {
			return err
		}
		// ak := os.Getenv("APP_ACCESS_KEY")
		// if ak != "" {
		// 	accessKey = ak
		// }

		secret, err := biz.GenerateAppSecret()

		if err != nil {
			return err
		}
		// sk := os.Getenv("APP_SECRET")
		// if sk != "" {
		// 	secret = sk
		// }
		appid := biz.GenerateAppid(snk)
		// if pid := os.Getenv("APPID"); pid != "" {
		// 	appid = pid
		// }
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
