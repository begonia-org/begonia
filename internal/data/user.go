package data

import (
	"context"
	"fmt"
	"time"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userRepoImpl struct {
	data  *Data
	local *LayeredCache
	cfg   *config.Config
}

func NewUserRepoImpl(data *Data, local *LayeredCache, cfg *config.Config) biz.UserRepo {
	return &userRepoImpl{data: data, local: local, cfg: cfg}
}

func (r *userRepoImpl) Add(ctx context.Context, user *api.Users) error {
	user.CreatedAt = timestamppb.Now()
	user.UpdatedAt = timestamppb.Now()
	err := r.data.Create(user)
	if err != nil {
		return err
	}

	return err
}
func (r *userRepoImpl) Get(ctx context.Context, key string) (*api.Users, error) {

	app := &api.Users{}
	err := r.data.Get(app, app, "uid = ? and is_deleted=0", key)
	if err != nil {
		return nil, err
	}
	err=tiga.DecryptStructAES([]byte(r.cfg.GetAesKey()), app, r.cfg.GetAesIv())
	if err!=nil{
		return nil,fmt.Errorf("decrypt user struct failed:%w",err)
	}
	return app, err
}


func (r *userRepoImpl) Del(ctx context.Context, key string) error {
	user, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("%s_%s", user.Name, time.Now().Format("20060102150405"))
	user.IsDeleted = true
	user.UpdatedAt = timestamppb.Now()
	user.Name = name
	user.Phone = fmt.Sprintf("%s_%s", user.Phone, time.Now().Format("20060102150405"))
	user.Email = fmt.Sprintf("%s_%s", user.Email, time.Now().Format("20060102150405"))
	user.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"is_deleted", "name"}}
	err = tiga.EncryptStructAES([]byte(r.cfg.GetAesKey()), user, r.cfg.GetAesIv())
	if err != nil {
		return fmt.Errorf("encrypt user struct failed:%w", err)
	}
	err = r.data.Update(ctx, user)
	return err
}
func (r *userRepoImpl) Patch(ctx context.Context, model *api.Users) error {
	err:=tiga.EncryptStructAES([]byte(r.cfg.GetAesKey()), model, r.cfg.GetAesIv())
	if err!=nil{
		return fmt.Errorf("encrypt patch user struct failed:%w",err)
	}
	return r.data.Update(ctx, model)
}
func (r *userRepoImpl) List(ctx context.Context, conds ...interface{}) ([]*api.Users, error) {
	apps := make([]*api.Users, 0)
	if err := r.data.List(&api.Users{}, &apps, conds...); err != nil {
		return nil, err
	}
	return apps, nil
}
