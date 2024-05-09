package biz

import (
    "context"
    "crypto/rand"
    "fmt"
    "strings"
    "time"

    "github.com/begonia-org/begonia/internal/pkg/config"
    "github.com/begonia-org/begonia/internal/pkg/errors"
    api "github.com/begonia-org/go-sdk/api/app/v1"
    common "github.com/begonia-org/go-sdk/common/api/v1"
    "github.com/spark-lence/tiga"
    "google.golang.org/grpc/codes"
    "google.golang.org/protobuf/types/known/timestamppb"
    "gorm.io/gorm"
)

type AppRepo interface {
    Add(ctx context.Context, apps *api.Apps) error
    Get(ctx context.Context, key string) (*api.Apps, error)
    Cache(ctx context.Context, prefix string, models *api.Apps, exp time.Duration) error
    Del(ctx context.Context, key string) error
    List(ctx context.Context, tags []string, status []api.APPStatus, page, pageSize int32) ([]*api.Apps, error)
    Patch(ctx context.Context, model *api.Apps) error
    GetSecret(ctx context.Context, accessKey string) (string, error)
    GetAppid(ctx context.Context, accessKey string) (string, error)
}

type AppUsecase struct {
    repo      AppRepo
    config    *config.Config
    snowflake *tiga.Snowflake
}

func NewAppUsecase(repo AppRepo, config *config.Config) *AppUsecase {
    sn, _ := tiga.NewSnowflake(1)
    return &AppUsecase{repo: repo, config: config, snowflake: sn}
}
func GenerateRandomString(n int) (string, error) {
    const lettersAndDigits = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, n)
    if _, err := rand.Read(b); err != nil {
        return "", fmt.Errorf("Failed to generate random string: %w", err)
    }

    for i := 0; i < n; i++ {
        // 将随机字节转换为lettersAndDigits中的一个有效字符
        b[i] = lettersAndDigits[b[i]%byte(len(lettersAndDigits))]
    }

    return string(b), nil
}
func (a *AppUsecase) newApp() *api.Apps {
    return &api.Apps{
        Status:    api.APPStatus_APP_ENABLED,
        IsDeleted: false,
    }

}
func (a *AppUsecase) CreateApp(ctx context.Context, in *api.AppsRequest, owner string) (*api.Apps, error) {
    // return a.repo.ListApps(ctx, conds...)
    appid := GenerateAppid(ctx, a.snowflake)
    accessKey, err := GenerateAppAccessKey(ctx)
    if err != nil {
        return nil, errors.New(err, int32(api.APPSvrCode_APP_CREATE_ERR), codes.Internal, "generate_app_access_key")

    }
    secret, err := GenerateAppSecret(ctx)
    if err != nil {
        return nil, errors.New(err, int32(api.APPSvrCode_APP_CREATE_ERR), codes.Internal, "generate_app_secret_key")
    }
    app := a.newApp()
    app.AccessKey = accessKey
    app.Secret = secret
    app.Appid = appid
    app.Name = in.Name
    app.Description = in.Description
    app.Tags = in.Tags
    app.Owner = owner
    err = a.Put(ctx, app, owner)
    if err != nil {
        return nil, err

    }
    return app, nil
}
func GenerateAppid(_ context.Context, snowflake *tiga.Snowflake) string {
    appid := snowflake.GenerateIDString()
    return appid
}
func GenerateAppAccessKey(_ context.Context) (string, error) {
    return GenerateRandomString(32)
}
func GenerateAppSecret(_ context.Context) (string, error) {
    return GenerateRandomString(64)
}

// AddApps 新增并缓存app
func (a *AppUsecase) Put(ctx context.Context, apps *api.Apps, owner string) (err error) {
    defer func() {
        if err != nil {
            // log.Println(err)
            if strings.Contains(err.Error(), "Duplicate entry") {
                err = errors.New(err, int32(api.APPSvrCode_APP_DUPLICATE_ERR), codes.AlreadyExists, "commit_app")
            } else {
                err = errors.New(err, int32(api.APPSvrCode_APP_CREATE_ERR), codes.Internal, "cache_apps")

            }
        }
    }()
    apps.Owner = owner
    err = a.repo.Add(ctx, apps)
    if err != nil {
        return err
    }
    prefix := a.config.GetAPPAccessKeyPrefix()
    err = a.repo.Cache(ctx, prefix, apps, time.Duration(0)*time.Second)
    return err
    // return a.repo.AddApps(ctx, apps)
}
func (a *AppUsecase) Get(ctx context.Context, key string) (*api.Apps, error) {
    app, err := a.repo.Get(ctx, key)
    if err != nil {
        return nil, errors.New(err, int32(api.APPSvrCode_APP_NOT_FOUND_ERR), codes.NotFound, "get_app")

    }
    return app, nil
}

func (a *AppUsecase) Cache(ctx context.Context, prefix string, models *api.Apps, exp time.Duration) error {
    return a.repo.Cache(ctx, prefix, models, exp)

}
func (a *AppUsecase) Del(ctx context.Context, key string) error {
    err := a.repo.Del(ctx, key)
    if err != nil {
        if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
            return errors.New(err, int32(api.APPSvrCode_APP_NOT_FOUND_ERR), codes.NotFound, "delete_app")
        }
        return errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "delete_app")

    }
    return nil
}
func (a *AppUsecase) Patch(ctx context.Context, in *api.AppsRequest, owner string) (*api.Apps, error) {
    app, err := a.Get(ctx, in.Appid)
    if err != nil {
        return nil, errors.New(err, int32(api.APPSvrCode_APP_NOT_FOUND_ERR), codes.NotFound, "get_app")
    }
    app.Name = in.Name
    app.Description = in.Description
    app.Tags = in.Tags
    app.UpdatedAt = timestamppb.Now()
    app.UpdateMask = in.UpdateMask
    err = a.repo.Patch(ctx, app)
    if err != nil {
        return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "update_app")
    }
    return app, nil
}

func (a *AppUsecase) List(ctx context.Context, in *api.AppsListRequest) ([]*api.Apps, error) {
    apps, err := a.repo.List(ctx, in.Tags, in.Status, in.Page, in.PageSize)
    if err != nil {
        return nil, errors.New(err, int32(api.APPSvrCode_APP_NOT_FOUND_ERR), codes.NotFound, "list_app")
    }
    return apps, nil
}
