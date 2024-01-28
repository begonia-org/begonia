package data

import (
	"context"

	common "github.com/begonia-org/begonia/common/api/v1"
	"github.com/begonia-org/begonia/internal/biz"
)

type fileRepoImpl struct {
	data *Data
}

func NewFileRepoImpl(data *Data) biz.FileRepo {
	return &fileRepoImpl{data: data}
}

// mysql
// AddFile(ctx context.Context, files []*common.Files) error
// DeleteFile(ctx context.Context, files []*common.Files) error
// UpdateFile(ctx context.Context, files []*common.Files) error
// GetFile(ctx context.Context, uri string) (*common.Files, error)
// ListFile(ctx context.Context, name []string) ([]*common.Files, error)
func (r *fileRepoImpl) AddFile(ctx context.Context, files []*common.Files) error {
	return nil
}

func (r *fileRepoImpl) DeleteFile(ctx context.Context, files []*common.Files) error {
	return nil
}

func (r *fileRepoImpl) UpdateFile(ctx context.Context, files []*common.Files) error {
	return nil
}

func (r *fileRepoImpl) GetFile(ctx context.Context, uri string) (*common.Files, error) {
	return nil, nil
}

func (r *fileRepoImpl) ListFile(ctx context.Context, name []string) ([]*common.Files, error) {
	return nil, nil
}
