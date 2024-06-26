package migrate

import (
	"context"
	"fmt"
	"time"

	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type UsersOperator struct {
	mysql *tiga.MySQLDao
}

func NewUsersOperator(mysql *tiga.MySQLDao) *UsersOperator {
	return &UsersOperator{mysql: mysql}
}
func (m *UsersOperator) InitAdminUser(passwd string, aseKey, ivKey string, name, email, phone string) (string, error) {
	userExist := &api.Users{}
	err := m.mysql.First(context.TODO(), userExist, "role = ? and is_deleted=0 and status=?", api.Role_ADMIN, api.USER_STATUS_ACTIVE)
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}
	if userExist.Uid == "" {
		snk, err := tiga.NewSnowflake(1)
		if err != nil {
			return "", err
		}
		// 初始化数据
		uid := snk.GenerateID()
		user := &api.Users{
			Uid:       fmt.Sprintf("%d", uid),
			Name:      name,
			Password:  passwd,
			Phone:     phone,
			Email:     email,
			Role:      api.Role_ADMIN,
			Dept:      "",
			Avatar:    "",
			Status:    api.USER_STATUS_ACTIVE,
			CreatedAt: timestamppb.New(time.Now()),
			UpdatedAt: timestamppb.New(time.Now()),
			IsDeleted: false,
		}

		err = tiga.EncryptStructAES([]byte(aseKey), user, ivKey)
		if err != nil {
			return "", err
		}
		err = m.mysql.Create(context.Background(), user)
		return user.Uid, err
	}
	return userExist.Uid, nil
}
