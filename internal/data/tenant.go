package data

import (
	"context"
	"fmt"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	api "github.com/begonia-org/go-sdk/api/user/v1"
	"github.com/spark-lence/tiga"
)

// type TenantRepo interface {
// 	Add(ctx context.Context, tenant *api.Tenants) error
// 	Get(ctx context.Context, key string) (*api.Tenants, error)
// 	Del(ctx context.Context, uidOrName string) error
// 	List(ctx context.Context, tags []string, status []api.USER_STATUS, page, pageSize int32) ([]*api.Tenants, error)
// 	Patch(ctx context.Context, model *api.Tenants) error
// }

type tenantRepoImpl struct {
	data *Data
	cfg  *config.Config
	curd biz.CURD
}

func (t *tenantRepoImpl) Add(ctx context.Context, tenant *api.Tenants) error {
	err := t.curd.Add(ctx, tenant, true, nil)
	return err
}
func (t *tenantRepoImpl) Get(ctx context.Context, key string) (*api.Tenants, error) {
	tenant := &api.Tenants{}
	err := t.curd.Get(ctx, tenant, false, "tenant_id=? or tenant_name=?", key, key)
	if err != nil || tenant.TenantId == "" {
		return nil, err
	}
	return tenant, nil
}
func (t *tenantRepoImpl) Del(ctx context.Context, tenantId string) error {
	return t.curd.Del(ctx, &api.Tenants{TenantId: tenantId}, false, nil)
}
func (t *tenantRepoImpl) List(ctx context.Context, tags []string, status []api.TENANTS_STATUS, page, pageSize int32) ([]*api.Tenants, error) {
	tenants := make([]*api.Tenants, 0)
	query := ""
	conds := make([]interface{}, 0)
	if len(tags) > 0 {
		query = "json_contains(json_array(?),tags)"
		conds = append(conds, tags)
	}
	if len(status) > 0 {
		if query == "" {
			query = "status in (?)"

		} else {
			query += " and status in (?)"
		}
		conds = append(conds, status)
	}
	pagination := &tiga.Pagination{
		Page:     page,
		PageSize: pageSize,
		Query:    query,
		Args:     conds,
	}
	err := t.curd.List(ctx, &tenants, pagination)
	if err != nil {
		return nil, err
	}
	return tenants, nil
}
func (t *tenantRepoImpl) Patch(ctx context.Context, model *api.Tenants) error {
	return t.curd.Update(ctx, model, false, nil)
}
func (t *tenantRepoImpl) AddBusiness(ctx context.Context, tenantBusiness *api.TenantsBusiness) error {
	tenant, err := t.Get(ctx, tenantBusiness.TenantId)
	if err != nil || tenant == nil || tenant.TenantId == "" {
		return fmt.Errorf("get tenant failed: %w or not found tenant before add business", err)
	}
	bs := &api.Business{BusinessId: tenantBusiness.BusinessId}
	err = t.curd.Get(ctx, bs, false, "business_id=?", tenantBusiness.BusinessId)
	if err != nil || bs.BusinessId == "" {
		return fmt.Errorf("get business failed: %w or not found business before add business", err)

	}
	return t.curd.Add(ctx, tenantBusiness, false, nil)
}

func (t *tenantRepoImpl) DelTenantBusiness(ctx context.Context, tenantId, businessId string) error {
	return t.curd.Del(ctx, &api.TenantsBusiness{TenantId: tenantId, BusinessId: businessId}, false, nil)
}
func (t *tenantRepoImpl) GetTenantBusiness(ctx context.Context, tenant, business string) (*api.TenantsBusiness,error) {
	tenantBusiness := &api.TenantsBusiness{}
	err := t.curd.Get(ctx, tenantBusiness, false, "(tenant_id=? or tenant_name=?) and (business_id=? or business_name=?)", tenant, tenant, business, business)
	if err != nil || tenantBusiness.BusinessId == "" || tenantBusiness.TenantId == ""{
		return nil, fmt.Errorf("get tenant business failed: %w or not found", err)
	}
	return tenantBusiness,nil
}
func (t *tenantRepoImpl)TenantBusinessList(ctx context.Context, tenantId string, page, pageSize int32) ([]*api.TenantsBusiness, error) {
	tenantBusinesses := make([]*api.TenantsBusiness, 0)
	pagination := &tiga.Pagination{
		Page:     page,
		PageSize: pageSize,
		Query:    "tenant_id=?",
		Args:     []interface{}{tenantId},
	}
	err := t.curd.List(ctx, &tenantBusinesses, pagination)
	if err != nil {
		return nil, err
	}
	return tenantBusinesses, nil
}
func NewTenantRepoImpl(data *Data, cfg *config.Config, curd biz.CURD) biz.TenantRepo {
	return &tenantRepoImpl{data: data, cfg: cfg, curd: curd}
}
