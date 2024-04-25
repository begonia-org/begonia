package data

import (
	"context"

	"github.com/begonia-org/begonia/internal/biz/file"
	common "github.com/begonia-org/go-sdk/common/api/v1"
)

type fileRepoImpl struct {
	data *Data
}

func NewFileRepoImpl(data *Data) file.FileRepo {
	return &fileRepoImpl{data: data}
}

// mysql
// AddFile(ctx context.Context, files []*common.Files) error
// DeleteFile(ctx context.Context, files []*common.Files) error
// UpdateFile(ctx context.Context, files []*common.Files) error
// GetFile(ctx context.Context, uri string) (*common.Files, error)
// ListFile(ctx context.Context, name []string) ([]*common.Files, error)
func (r *fileRepoImpl) UploadFile(ctx context.Context, files []*common.Files) error {
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
