package biz

import (
	"context"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/redis/go-redis/v9"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc/codes"
)

type UserRepo interface {
	Add(ctx context.Context, apps *api.Users) error
	Get(ctx context.Context, key string) (*api.Users, error)
	Del(ctx context.Context, key string) error
	List(ctx context.Context, dept []string, status []api.USER_STATUS, page, pageSize int32) ([]*api.Users, error)
	Patch(ctx context.Context, model *api.Users) error
	Cache(ctx context.Context, prefix string, models []*api.Users, exp time.Duration, getValue func(user *api.Users) ([]byte, interface{})) redis.Pipeliner
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

	err = u.repo.Add(ctx, users)
	return

}
func (u *UserUsecase) Get(ctx context.Context, key string) (*api.Users, error) {
	user, err := u.repo.Get(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.New(err, int32(api.UserSvrCode_USER_NOT_FOUND_ERR), codes.NotFound, "get_user")

		}
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_user")
	}
	return user, nil
}

func (u *UserUsecase) Update(ctx context.Context, model *api.Users) error {
	err := u.repo.Patch(ctx, model)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.New(err, int32(api.UserSvrCode_USER_NOT_FOUND_ERR), codes.NotFound, "get_user")

		}
		return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_user")
	}
	return nil
}
func (u *UserUsecase) Delete(ctx context.Context, uid string) error {
	err := u.repo.Del(ctx, uid)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.New(err, int32(api.UserSvrCode_USER_NOT_FOUND_ERR), codes.NotFound, "get_user")

		}
		return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_user")
	}
	return nil
}
