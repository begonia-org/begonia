package biz

import (
	"context"
	"fmt"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/config"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc/codes"
)

type TenantRepo interface {
	Add(ctx context.Context, tenant *api.Tenants) error
	Get(ctx context.Context, key string) (*api.Tenants, error)
	Del(ctx context.Context, uidOrName string) error
	List(ctx context.Context, tags []string, status []api.TENANTS_STATUS, page, pageSize int32) ([]*api.Tenants, error)
	Patch(ctx context.Context, model *api.Tenants) error
	AddBusiness(ctx context.Context, tenantBusiness *api.TenantsBusiness) error
	DelTenantBusiness(ctx context.Context, tenantId, businessId string) error
	GetTenantBusiness(ctx context.Context, tenant, business string) (*api.TenantsBusiness, error)
	TenantBusinessList(ctx context.Context, tenantId string, page, pageSize int32) ([]*api.TenantsBusiness, error)
}

type TenantUsecase struct {
	repo      TenantRepo
	business  *BusinessUsecase
	cfg       *config.Config
	snowflake *tiga.Snowflake
}

func NewTenantUsecase(repo TenantRepo, business *BusinessUsecase, cfg *config.Config) *TenantUsecase {
	snk, _ := tiga.NewSnowflake(1)
	return &TenantUsecase{repo: repo, business: business, cfg: cfg, snowflake: snk}
}

func (u *TenantUsecase) Get(ctx context.Context, uidOrName string) (*api.Tenants, error) {
	user, err := u.repo.Get(ctx, uidOrName)
	if err != nil || user == nil || user.TenantId == "" {
		return nil, gosdk.NewError(fmt.Errorf("get tenant fail:%w or not found tenant", err), int32(api.UserSvrCode_USER_NOT_FOUND_ERR), codes.NotFound, "get_tenant")
	}
	return user, nil
}
func (u *TenantUsecase) Add(ctx context.Context, in *api.PostTenantRequest, createdBy string) (*api.Tenants, error) {
	tenants := &api.Tenants{
		TenantName:  in.TenantName,
		Email:       in.Email,
		Description: in.Description,
		Tags:        in.Tags,
		TenantId:    u.snowflake.GenerateIDString(),
		CreatedBy:   createdBy,
	}
	err := u.repo.Add(ctx, tenants)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, gosdk.NewError(err, int32(api.UserSvrCode_USER_USERNAME_DUPLICATE_ERR), codes.AlreadyExists, "add_tenant")
		}
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "add_tenant")
	}
	return tenants, nil
}

func (u *TenantUsecase) Update(ctx context.Context, in *api.PatchTenantRequest) (*api.Tenants, error) {
	tenant := &api.Tenants{
		TenantId:    in.TenantId,
		TenantName:  in.TenantName,
		Email:       in.Email,
		Description: in.Description,
		Tags:        in.Tags,
		Status:      in.Status,
		AdminId:     in.AdminId,
		UpdateMask:  in.UpdateMask,
	}

	err := u.repo.Patch(ctx, tenant)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, gosdk.NewError(fmt.Errorf("Update tenant error:%w", err), int32(common.Code_CONFLICT), codes.AlreadyExists, "patch_app")
		}
		if strings.Contains(err.Error(), "not found") {
			return nil, gosdk.NewError(fmt.Errorf("Update tenant error:%w", err), int32(api.UserSvrCode_USER_NOT_FOUND_ERR), codes.NotFound, "get_user")
		}

		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_user")
	}
	return u.Get(ctx, in.TenantId)
}
func (u *TenantUsecase) Delete(ctx context.Context, uidOrName string) error {
	err := u.repo.Del(ctx, uidOrName)
	if err != nil {
		return gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_user")
	}
	return nil
}
func (u *TenantUsecase) List(ctx context.Context, tags []string, status []api.TENANTS_STATUS, page, pageSize int32) ([]*api.Tenants, error) {
	tenants, err := u.repo.List(ctx, tags, status, page, pageSize)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_tenants")

	}
	return tenants, nil
}
func (t *TenantUsecase) AddTenantBusiness(ctx context.Context, tenantId, businessId string, plan, createdBy string) (*api.TenantsBusiness, error) {
	tenant, err := t.Get(ctx, tenantId)
	if err != nil || tenant == nil || tenant.TenantId == "" {
		return nil, gosdk.NewError(fmt.Errorf("get tenant fail:%w or tenant not found", err), int32(api.UserSvrCode_USER_NOT_FOUND_ERR), codes.InvalidArgument, "get_tenant")
	}
	business, err := t.business.Get(ctx, businessId)
	if err != nil || business == nil || business.BusinessId == "" {
		return nil, gosdk.NewError(fmt.Errorf("get business fail:%w or business not found", err), int32(api.UserSvrCode_USER_NOT_FOUND_ERR), codes.InvalidArgument, "get_business")
	}
	tb := &api.TenantsBusiness{
		TenantId:     tenantId,
		BusinessId:   businessId,
		Plan:         plan,
		TenantName:   tenant.TenantName,
		BusinessName: business.BusinessName,
		CreatedBy:    createdBy,
	}
	err = t.repo.AddBusiness(ctx, tb)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "add_tenants_business")

	}
	return tb, nil
}
func (t *TenantUsecase) GetTenantBusiness(ctx context.Context, tenant, business string) (*api.TenantsBusiness, error) {
	tb, err := t.repo.GetTenantBusiness(ctx, tenant, business)
	if err != nil || tb == nil || tb.TenantId == "" {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_tenants_business")
	}
	return tb, nil
}
func (t *TenantUsecase) DelTenantBusiness(ctx context.Context, tenantId, businessId string) error {
	err := t.repo.DelTenantBusiness(ctx, tenantId, businessId)
	if err != nil {
		return gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "del_tenants_business")

	}
	return nil
}

func (t *TenantUsecase) ListTenantBusiness(ctx context.Context, tenantId string, page, pageSize int32) ([]*api.TenantsBusiness, error) {
	tbs, err := t.repo.TenantBusinessList(ctx, tenantId, page, pageSize)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_tenants_business")
	}
	return tbs, nil
}
