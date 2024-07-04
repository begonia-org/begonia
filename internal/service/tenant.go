package service

import (
	"context"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/config"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	user "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/begonia-org/go-sdk/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type TenantService struct {
	api.UnimplementedTenantsServiceServer
	tenant *biz.TenantUsecase
	cfg    *config.Config
	log    logger.Logger
}

func NewTenantService(tenant *biz.TenantUsecase, cfg *config.Config, log logger.Logger) api.TenantsServiceServer {
	return &TenantService{tenant: tenant, cfg: cfg, log: log}
}
func (t *TenantService) Register(ctx context.Context, in *api.PostTenantRequest) (*api.Tenants, error) {
	identity := GetIdentity(ctx)
	if identity == "" {
		return nil, gosdk.NewError(pkg.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")

	}
	return t.tenant.Add(ctx, in, identity)

}
func (t *TenantService) Get(ctx context.Context, in *api.GetTenantRequest) (*api.Tenants, error) {
	tenant, err := t.tenant.Get(ctx, in.Tenant)
	return tenant, err
}
func (t *TenantService) Update(ctx context.Context, in *api.PatchTenantRequest) (*api.Tenants, error) {

	return t.tenant.Update(ctx, in)

}
func (t *TenantService) List(ctx context.Context,in *api.ListTenantsRequest) (*api.ListTenantsResponse, error) {
	tenants,err:=t.tenant.List(ctx, in.Tags, in.Status, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	
	}
	return &api.ListTenantsResponse{Tenants: tenants}, nil

}
func (t *TenantService) Delete(ctx context.Context, in *api.DeleteTenantRequest) (*api.DeleteTenantResponse, error) {
	err := t.tenant.Delete(ctx, in.Tenant)
	if err != nil {
		return nil, err
	}
	return &api.DeleteTenantResponse{}, nil
}
func (t *TenantService) AddTenantBusiness(ctx context.Context, in *api.AddTenantBusinessRequest) (*api.TenantsBusiness, error) {
	identity := GetIdentity(ctx)
	if identity == "" {
		return nil, gosdk.NewError(pkg.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")

	}
	tb, err := t.tenant.AddTenantBusiness(ctx, in.TenantId, in.BusinessId, in.Plan, identity)
	return tb, err

}
func (t *TenantService) DeleteTenantBusiness(ctx context.Context, in *api.DeleteTenantBusinessRequest) (*api.DeleteTenantBusinessResponse, error) {
	err := t.tenant.DelTenantBusiness(ctx, in.TenantId, in.BusinessId)
	if err != nil {
		return nil, err
	}
	return &api.DeleteTenantBusinessResponse{}, nil
}
func (t *TenantService) ListTenantBusiness(ctx context.Context, in *api.ListTenantBusinessRequest) (*api.ListTenantBusinessResponse, error) {
	tb, err := t.tenant.ListTenantBusiness(ctx, in.Tenant, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &api.ListTenantBusinessResponse{TenantsBusiness: tb}, nil
}
func (b *TenantService) Desc() *grpc.ServiceDesc {
	return &api.TenantsService_ServiceDesc
}
