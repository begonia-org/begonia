package data

import "context"
import 	api "github.com/begonia-org/begonia/api/v1"

type AppRepoImpl struct {
	data *Data
}

func NewAppRepoImpl(data *Data) *AppRepoImpl {
	return &AppRepoImpl{data: data}
}

func (r *AppRepoImpl) AddApps(ctx context.Context, apps []*api.Apps) error {
	sources := NewSourceTypeArray(apps)
	return r.data.CreateInBatches(sources)
}
func (r *AppRepoImpl) GetApps(ctx context.Context, keys []string) ([]*api.Apps, error) {
	apps  := make([]*api.Apps, 0)
	err := r.data.List(&api.Apps{}, &apps, "appid in (?) or access_key in (?)", keys)
	return apps, err
}
