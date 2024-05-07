package migrate

import (
	"context"
	"log"
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
func (m *APPOperator) InitAdminAPP(owner string) error {
	app := &api.Apps{}
	err := m.mysql.First(app, "name = ?", "admin-app")
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if app.Appid == "" {
		snk, err := tiga.NewSnowflake(1)
		if err != nil {
			return err
		}
		// 初始化数据
		// uid := snk.GenerateID()
		accessKey, err := biz.GenerateAppAccessKey(context.TODO())
		if err != nil {
			return err
		}
		secret, err := biz.GenerateAppSecret(context.TODO())
		if err != nil {
			return err
		}
		app = &api.Apps{
			Appid:       biz.GenerateAppid(context.TODO(), snk),
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
	log.Print("########################################admin-app###############################")
	log.Printf("Init appid:%s", app.Appid)
	log.Printf("Init accessKey:%s", app.AccessKey)
	log.Printf("Init secret:%s", app.Secret)
	log.Print("#################################################################################")
	return nil
}
