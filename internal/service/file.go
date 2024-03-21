package service

import (
	"context"

	api "github.com/begonia-org/begonia/api/v1"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
)

type FileService struct {
	api.UnimplementedFileServiceServer
	biz    *biz.FileUsecase
	config *config.Config
}

func NewFileService(biz *biz.FileUsecase, config *config.Config) *FileService {
	return &FileService{biz: biz, config: config}
}

func (f *FileService) Upload(ctx context.Context, in *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	return f.biz.Upload(ctx, in)
}

func (f *FileService) InitiateMultipartUpload(ctx context.Context, in *api.InitiateMultipartUploadRequest) (*api.InitiateMultipartUploadResponse, error) {
	return f.biz.InitiateUploadFile(ctx, in)
}
func (f *FileService) UploadMultipartFile(ctx context.Context, in *api.UploadMultipartFileRequest) (*api.UploadMultipartFileResponse, error) {
	return f.biz.UploadMultipartFileFile(ctx, in)
}
func (f *FileService) CompleteMultipartUpload(ctx context.Context, in *api.CompleteMultipartUploadRequest) (*api.CompleteMultipartUploadResponse, error) {
	return f.biz.CompleteMultipartUploadFile(ctx, in)
}
func (f *FileService) AbortMultipartUpload(ctx context.Context, in *api.AbortMultipartUploadRequest) (*api.AbortMultipartUploadResponse, error) {
	return f.biz.AbortMultipartUpload(ctx, in)
}
func (f *FileService) Download(ctx context.Context, in *api.DownloadRequest) (*httpbody.HttpBody, error) {
	return f.biz.Download(ctx, in)
}
func (f *FileService) Delete(ctx context.Context, in *api.DeleteRequest) (*api.DeleteResponse, error) {
	return f.biz.Delete(ctx, in)
}
func (f *FileService) Desc() *grpc.ServiceDesc {
	return &api.FileService_ServiceDesc
}
