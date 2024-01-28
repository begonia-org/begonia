package biz

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	common "github.com/begonia-org/begonia/common/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FileRepo interface {
	// mysql
	AddFile(ctx context.Context, files []*common.Files) error
	DeleteFile(ctx context.Context, files []*common.Files) error
	UpdateFile(ctx context.Context, files []*common.Files) error
	GetFile(ctx context.Context, uri string) (*common.Files, error)
	ListFile(ctx context.Context, name []string) ([]*common.Files, error)
}

type FileUsecase struct {
	repo   FileRepo
	config *config.Config
}

func NewFileUsecase(repo FileRepo, config *config.Config) *FileUsecase {
	return &FileUsecase{repo: repo, config: config}
}

func (f *FileUsecase) Upload(uploadDir string, filename string, stream common.FileService_UploadFileServer) (*common.Files, error) {
	// var once sync.Once
	var file *os.File
	var err error
	hash := sha256.New()

	var filePath string

	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	// 确保目录存在
	if err = os.MkdirAll(uploadDir, 0755); err != nil {
		logger.Logger.Errorf("Failed to create directory: %v", err)
		err = fmt.Errorf("Failed to create directory: %w", err)
	}

	// 创建文件
	filePath = filepath.Join(uploadDir, filename)
	filename = filePath
	file, err = os.Create(filename)
	if err != nil {
		logger.Logger.Errorf("Failed to create file: %s", err.Error())
		err = fmt.Errorf("Failed to create file: %w", err)
	}
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			// 文件接收完毕
			break
		}
		if err != nil {
			logger.Logger.Errorf("Error while receiving chunk: %v", err)
			return nil, fmt.Errorf("Error while receiving chunk: %w", err)
		}

		if file != nil {
			// 向文件写入数据
			if _, err = file.Write(chunk.Data); err != nil {
				logger.Logger.Errorf("Failed to write to file: %s", err.Error())
				return nil, fmt.Errorf("Failed to write to file: %w", err)
			}
			// 更新哈希值
			if _, err = hash.Write(chunk.Data); err != nil {
				logger.Logger.Errorf("Failed to update hash: %v", err)
				return nil, fmt.Errorf("Failed to update hash: %w", err)
			}
		}
	}
	if err != nil {
		return nil, err

	}
	uploadRootDir := f.config.GetUploadDir()
	uri, err := filepath.Rel(uploadRootDir, filePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to get file uri: %w", err)

	}
	// 计算最终的哈希值
	hashInBytes := hash.Sum(nil)
	sha256Hash := fmt.Sprintf("%x", hashInBytes)
	return &common.Files{
		Uri:       uri,
		Name:      filepath.Ext(filepath.Base(filePath)),
		Sha:       sha256Hash,
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}, err

}

func (f *FileUsecase) AddFile(ctx context.Context, files []*common.Files) error {
	return f.repo.AddFile(ctx, files)
}
