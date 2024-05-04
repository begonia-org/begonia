package biz

import (
	"context"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc/codes"
)

type UserRepo interface {
	Add(ctx context.Context, apps *api.Users) error
	Get(ctx context.Context, key string) (*api.Users, error)
	Del(ctx context.Context, key string) error
	List(ctx context.Context, conds ...interface{}) ([]*api.Users, error)
	Patch(ctx context.Context, model *api.Users) error
}

type UserUsecase struct {
	repo      UserRepo
	config    *config.Config
	snowflake *tiga.Snowflake
}

func NewUserUsecase(repo UserRepo, config *config.Config) *UserUsecase {
	sn, _ := tiga.NewSnowflake(1)

	return &UserUsecase{repo: repo, config: config, snowflake: sn}
}

func (u *UserUsecase) Add(ctx context.Context, users *api.Users) (err error) {
	defer func() {
		if err != nil {
			// log.Println(err)
			if strings.Contains(err.Error(), "Duplicate entry") {
				err = errors.New(err, int32(api.UserSvrCode_USER_USERNAME_DUPLICATE_ERR), codes.AlreadyExists, "commit_app")
			} else {
				err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "cache_apps")

			}
		}
	}()
	users.Uid = u.snowflake.GenerateIDString()
	ivKey := u.config.GetAesIv()
	aseKey := u.config.GetAesKey()

	err = tiga.EncryptStructAES([]byte(aseKey), users, ivKey)
	if err != nil {
		return
	}
	err = u.repo.Add(ctx, users)
	return

}
func (u *UserUsecase) Get(ctx context.Context, key string) (*api.Users, error) {
	return u.repo.Get(ctx, key)
}

func (u *UserUsecase) Update(ctx context.Context, model *api.Users) error {
	return u.repo.Patch(ctx, model)
}
func (u *UserUsecase) Delete(ctx context.Context, uid string) error {
	return u.repo.Del(ctx, uid)
}
