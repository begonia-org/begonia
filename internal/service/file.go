package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/begonia-org/begonia/common"
	api "github.com/begonia-org/begonia/common/api/v1"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
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

func (f *FileService) UploadFile(stream api.FileService_UploadFileServer) error {
	uploadDir := f.config.GetUploadDir()
	t := time.Now().UnixMilli()

	header, err := common.NewHeadersFromContext(stream.Context())
	if err != nil {
		return err
	}

	filename := header.GetHeader("x-filename")
	if filename == "" {
		return fmt.Errorf("filename is empty")
	}
	uid := header.GetHeader("x-uid")
	dir := filepath.Join(uploadDir, uid, fmt.Sprintf("%d", t))
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	file, err := f.biz.Upload(dir, filename, stream)
	if err != nil {
		return err
	}
	files := []*api.Files{file}
	err = f.biz.AddFile(stream.Context(), files)
	if err != nil {
		return err
	}
	return stream.SendAndClose(&api.UploadAPIResponse{
		Uri:    file.Uri,
		Sha256: file.Sha,
	})
}

func (f *FileService) Desc() *grpc.ServiceDesc {
	return &api.FileService_ServiceDesc
}
