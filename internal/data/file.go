package data

import (
	"context"
	"fmt"
	"log"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/biz/file"
	api "github.com/begonia-org/go-sdk/api/file/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/spark-lence/tiga"
)

type fileRepoImpl struct {
	data *Data
	curd biz.CURD
}

func NewFileRepoImpl(data *Data, curd biz.CURD) file.FileRepo {
	return &fileRepoImpl{data: data, curd: curd}
}

func (f *fileRepoImpl) UpsertFile(ctx context.Context, in *api.Files) (bool, error) {
	in.UpdatedAt = timestamppb.Now()
	mask := make([]string, 0)
	if in.UpdateMask != nil {
		mask = in.UpdateMask.Paths
	}
	// log.Printf("mask:%v", in.Uid)
	return f.data.db.Upsert(ctx, in, mask...)
}
func (f *fileRepoImpl) DelFile(ctx context.Context, engine, bucket, key string) error {
	// return f.curd.Del(ctx, &api.Files{Uid: fid},false)
	return f.data.db.UpdateSelectColumns(ctx, &api.Files{Engine: engine, Bucket: bucket, Key: key}, &api.Files{IsDeleted: true}, "is_deleted")
}
func (f *fileRepoImpl) UpsertBucket(ctx context.Context, bucket *api.Buckets) (bool, error) {
	bucket.UpdatedAt = timestamppb.Now()
	mask := make([]string, 0)
	if bucket.UpdateMask != nil {
		mask = bucket.UpdateMask.Paths
	}
	return f.data.db.Upsert(ctx, bucket, mask...)
}
func (f *fileRepoImpl) DelBucket(ctx context.Context, bucketId string) error {
	return f.curd.Del(ctx, &api.Buckets{Uid: bucketId}, false)
}
func (f *fileRepoImpl) GetFileById(ctx context.Context, fid string) (*api.Files, error) {
	file := &api.Files{Uid: fid}
	err := f.curd.Get(ctx, file, false, "uid = ?", fid)
	if err != nil {
		return nil, err
	}
	return file, nil
}
func (f *fileRepoImpl) GetFile(ctx context.Context, engine, bucket, key string) (*api.Files, error) {
	file := &api.Files{}
	err := f.curd.Get(ctx, file, false, "`file_key`=? and `bucket`=? and `fs_engine`=?", key, bucket, engine)
	if err != nil {
		return nil, err
	}
	return file, nil
}
func (f *fileRepoImpl) List(ctx context.Context, page, pageSize int32, bucket, engine, owner string) ([]*api.Files, error) {
	files := make([]*api.Files, 0)
	pagination := &tiga.Pagination{
		Page:     page,
		PageSize: pageSize,
		Args:     make([]interface{}, 0),
	}
	pagination.Query = ""
	if bucket != "" {
		pagination.Query = "bucket = ?"
		pagination.Args = append(pagination.Args, bucket)
	}
	if engine != "" {
		if pagination.Query != "" {
			pagination.Query = fmt.Sprintf("%s and fs_engine = ?", pagination.Query)
		} else {
			pagination.Query = "fs_engine = ?"
		}
		pagination.Args = append(pagination.Args, engine)
	}
	log.Printf("query:%s,args:%s", pagination.Query, pagination.Args)
	err := f.curd.List(ctx, &files, pagination)

	if err != nil {
		return nil, err
	}
	return files, nil
}
