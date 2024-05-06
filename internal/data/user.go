package data

import (
    "context"
    "fmt"
    "time"

    "github.com/begonia-org/begonia/internal/biz"
    "github.com/begonia-org/begonia/internal/pkg/config"
    api "github.com/begonia-org/go-sdk/api/user/v1"
    "github.com/redis/go-redis/v9"
    "github.com/spark-lence/tiga"
)

type userRepoImpl struct {
    data  *Data
    local *LayeredCache
    cfg   *config.Config
    curd  biz.CURD
}

func NewUserRepoImpl(data *Data, local *LayeredCache, curd biz.CURD, cfg *config.Config) biz.UserRepo {
    return &userRepoImpl{data: data, local: local, cfg: cfg, curd: curd}
}

func (r *userRepoImpl) Add(ctx context.Context, user *api.Users) error {

    err := r.curd.Add(ctx, user, true)
    return err
}
func (r *userRepoImpl) Get(ctx context.Context, key string) (*api.Users, error) {

    app := &api.Users{}
    err := r.curd.Get(ctx, app, true, "uid = ?", key)
    if err != nil||app.Uid=="" {
        return nil, fmt.Errorf("get user failed: %w", err)
    }
    return app, err
}

func (r *userRepoImpl) Del(ctx context.Context, key string) error {
    user, err := r.Get(ctx, key)
    if err != nil {
        return err
    }
    err = r.curd.Del(ctx, user, true)
    return err
}
func (r *userRepoImpl) Patch(ctx context.Context, model *api.Users) error {

    return r.curd.Update(ctx, model, true)
}
func (r *userRepoImpl) List(ctx context.Context, dept []string, status []api.USER_STATUS, page, pageSize int32) ([]*api.Users, error) {
    apps := make([]*api.Users, 0)
    query := ""
    conds := make([]interface{}, 0)
    if len(dept) > 0 {
        query = "dept in (?)"
        conds = append(conds, dept)
    }
    if len(status) > 0 {
        if query != "" {
            query += " and "
        }
        query += "status in (?)"
        conds = append(conds, status)
    }
    pagination := &tiga.Pagination{
        Page:     page,
        PageSize: pageSize,
        Query:    query,
        Args:     conds,
    }
    if err := r.curd.List(ctx, &apps, pagination); err != nil {
        return nil, err
    }

    return apps, nil
}

func (u *userRepoImpl) Cache(ctx context.Context, prefix string, models []*api.Users, exp time.Duration, getValue func(user *api.Users) ([]byte, interface{})) redis.Pipeliner {
    kv := make([]interface{}, 0)
    for _, model := range models {
        // status,_:=tiga.IntToBytes(int(model.Status))
        valByte, val := getValue(model)
        key := fmt.Sprintf("%s:%s", prefix, model.Uid)
        if err := u.cacheUsers(ctx, prefix, model.Uid, valByte, exp); err != nil {
            return nil
        }
        kv = append(kv, key, val)
    }
    return u.data.BatchCacheByTx(ctx, exp, kv...)
}
func (u *userRepoImpl) cacheUsers(ctx context.Context, prefix string, uid string, value []byte, exp time.Duration) error {
    key := fmt.Sprintf("%s:%s", prefix, uid)
    err := u.local.Set(ctx, key, value, exp)
    if err != nil {
        return err
    }
    return nil
}
