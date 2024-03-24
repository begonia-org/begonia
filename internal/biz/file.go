package biz

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	api "github.com/begonia-org/begonia/api/v1"
	common "github.com/begonia-org/begonia/common/api/v1"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc/codes"
)

type FileRepo interface {
	// mysql
	UploadFile(ctx context.Context, files []*common.Files) error
	DeleteFile(ctx context.Context, files []*common.Files) error
	UpdateFile(ctx context.Context, files []*common.Files) error
	GetFile(ctx context.Context, uri string) (*common.Files, error)
	ListFile(ctx context.Context, name []string) ([]*common.Files, error)
}

type FileUsecase struct {
	repo      FileRepo
	config    *config.Config
	snowflake *tiga.Snowflake
}

func NewFileUsecase(repo FileRepo, config *config.Config) *FileUsecase {
	snk, _ := tiga.NewSnowflake(1)
	return &FileUsecase{repo: repo, config: config, snowflake: snk}
}
func (f *FileUsecase) getPartsDir(key string) string {
	return filepath.Join(f.config.GetUploadDir(), key, "parts")
}
func (f *FileUsecase) spiltKey(key string) (string, string, error) {
	if strings.HasPrefix(key, "/") {
		return "", "", errors.New(errors.ErrInvalidFileKey, int32(api.FileSvrStatus_FILE_INVAILDATE_KEY_ERR), codes.InvalidArgument, "invalid_key")
	}
	if strings.Contains(key, "/") {
		name := filepath.Base(key)
		// filename := getFilenameWithoutExt(name)
		return filepath.Dir(key), name, nil
	}
	return "./", key, nil
}
func (f *FileUsecase) InitiateUploadFile(ctx context.Context, in *api.InitiateMultipartUploadRequest) (*api.InitiateMultipartUploadResponse, error) {
	uploadId := f.snowflake.GenerateIDString()
	_, _, err := f.spiltKey(in.Key)
	if err != nil {
		return nil, err
	}
	saveDir := f.getPartsDir(uploadId)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_upload_dir")
		return nil, err
	}
	return &api.InitiateMultipartUploadResponse{
		UploadId: uploadId,
	}, nil
}

// 文件是否存在
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
func getFilenameWithoutExt(filename string) string {
	return filename[:len(filename)-len(filepath.Ext(filename))]
}
func getSHA256(data []byte) string {
	hash := sha256.Sum256([]byte(data))

	// 将哈希值格式化为十六进制字符串
	hashHex := fmt.Sprintf("%x", hash)
	return hashHex
}
func (f *FileUsecase) Upload(ctx context.Context, in *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	if in.Key == "" {
		return nil, errors.New(errors.ErrInvalidFileKey, int32(api.FileSvrStatus_FILE_INVAILDATE_KEY_ERR), codes.InvalidArgument, "invalid_key")
	}

	subDir, filename, err := f.spiltKey(in.Key)
	if err != nil {
		return nil, err
	}
	saveDir := filepath.Join(f.config.GetUploadDir(), subDir)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_upload_dir")
	}
	filePath := filepath.Join(saveDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_file")
	}
	defer file.Close()
	_, err = file.Write(in.Content)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "write_file")
	}
	uri, err := f.getUri(filePath)
	if err != nil {
		return nil, err
	}
	sha256Hash := getSHA256(in.Content)
	if sha256Hash != in.Sha256 {
		os.Remove(filePath)
		err := errors.New(errors.ErrSHA256NotMatch, int32(api.FileSvrStatus_FILE_SHA256_NOT_MATCH_ERR), codes.InvalidArgument, "sha256_not_match")
		return nil, err

	}
	return &api.UploadFileResponse{
		Uri: uri,
	}, nil

}
func (f *FileUsecase) UploadMultipartFileFile(ctx context.Context, in *api.UploadMultipartFileRequest) (*api.UploadMultipartFileResponse, error) {

	if in.UploadId == "" {
		return nil, errors.New(errors.ErrUploadIdMissing, int32(api.FileSvrStatus_FILE_UPLOADID_MISSING_ERR), codes.InvalidArgument, "upload_id_not_found")
	}
	if in.PartNumber == 0 {
		return nil, errors.New(errors.ErrPartNumberMissing, int32(api.FileSvrStatus_FILE_PARTNUMBER_MISSING_ERR), codes.InvalidArgument, "part_number_not_found")

	}
	// 分片上传处理
	uploadId := in.UploadId

	saveDir := f.getPartsDir(uploadId)
	if !pathExists(saveDir) {
		err := errors.New(errors.ErrUploadNotInitiate, int32(api.FileSvrStatus_FILE_UPLOAD_NOT_INITIATE_ERR), codes.NotFound, "upload_dir_not_found")
		return nil, err
	}

	filePath := filepath.Join(saveDir, fmt.Sprintf("%08d.part", in.PartNumber))

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_file")
	}
	defer file.Close()
	// 写入文件
	_, err = file.Write(in.Content)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "write_file")

	}
	// 计算哈希值
	sha256Hash := getSHA256(in.Content)
	if sha256Hash != in.Sha256 {
		os.Remove(filePath)
		err := errors.New(errors.ErrSHA256NotMatch, int32(api.FileSvrStatus_FILE_SHA256_NOT_MATCH_ERR), codes.InvalidArgument, "sha256_not_match")
		return nil, err

	}
	uri, err := f.getUri(filePath)
	if err != nil {
		return nil, err
	}
	return &api.UploadMultipartFileResponse{
		Uri: uri,
	}, nil
}
func (f *FileUsecase) getSortedFiles(dirPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".part") { // 假设文件是.txt格式
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 按文件名排序
	sort.Strings(files)

	return files, nil
}
func (f FileUsecase) mergeFiles(files []string, outputFile string) error {
	baseDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return err
	}
	out, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, file := range files {
		in, err := os.Open(file)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, in)
		in.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
func (f *FileUsecase) mvDir(src, dst string) error {
	newPath := filepath.Join(dst, filepath.Base(src))
	if _, err := os.Stat(newPath); err == nil {
		// 目标路径存在，尝试删除
		if err := os.RemoveAll(newPath); err != nil {
			return err
		}
	}

	return os.Rename(src, newPath)
}
func (f *FileUsecase) getPersistenceKeyParts(key string) string {
	if strings.Contains(key, ".") {
		key = key[:strings.LastIndex(key, ".")]

	}
	return filepath.Join(f.config.GetUploadDir(), "parts", key)
}
func (f *FileUsecase) getUri(filePath string) (string, error) {
	uploadRootDir := f.config.GetUploadDir()
	uri, err := filepath.Rel(uploadRootDir, filePath)
	if err != nil {
		return "", errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_file_uri")
	}
	return uri, nil

}
func (f *FileUsecase) AbortMultipartUpload(ctx context.Context, in *api.AbortMultipartUploadRequest) (*api.AbortMultipartUploadResponse, error) {
	partsDir := f.getPartsDir(in.UploadId)
	if !pathExists(partsDir) {
		err := errors.New(errors.ErrUploadIdNotFound, int32(api.FileSvrStatus_FILE_NOT_FOUND_UPLOADID_ERR), codes.NotFound, "upload_id_not_found")
		return nil, err

	}
	err := os.RemoveAll(partsDir)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "remove_parts_dir")
	}
	return &api.AbortMultipartUploadResponse{}, nil
}
func (f *FileUsecase) CompleteMultipartUploadFile(ctx context.Context, in *api.CompleteMultipartUploadRequest) (*api.CompleteMultipartUploadResponse, error) {

	if in.UploadId == "" {
		err := errors.New(errors.ErrUploadIdMissing, int32(api.FileSvrStatus_FILE_UPLOADID_MISSING_ERR), codes.InvalidArgument, "upload_id_not_found")
		return nil, err
	}
	partsDir := f.getPartsDir(in.UploadId)
	if !pathExists(partsDir) {
		err := errors.New(errors.ErrUploadIdNotFound, int32(api.FileSvrStatus_FILE_NOT_FOUND_UPLOADID_ERR), codes.NotFound, "upload_id_not_found")
		return nil, err

	}
	files, err := f.getSortedFiles(partsDir)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_sorted_files")

	}
	subDir, filename, err := f.spiltKey(in.Key)
	if err != nil {
		return nil, err
	}
	saveDir := filepath.Join(f.config.GetUploadDir(), subDir)
	filePath := filepath.Join(saveDir, filename)
	// merge files to uploadDir/key
	err = f.mergeFiles(files, filePath)
	if err != nil {
		return nil, errors.New(fmt.Errorf("merge file error"), int32(common.Code_INTERNAL_ERROR), codes.Internal, "merge_files")
	}
	// the parts file has been merged, remove the parts dir to uploadDir/parts/key
	keyParts := f.getPersistenceKeyParts(in.Key)
	if err = os.MkdirAll(keyParts, 0755); err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_parts_dir")
	}
	err = f.mvDir(partsDir, keyParts)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "mv_dir")

	}
	uri, err := f.getUri(filePath)
	if err != nil {
		return nil, err

	}
	os.RemoveAll(filepath.Join(f.config.GetUploadDir(), in.UploadId))
	return &api.CompleteMultipartUploadResponse{
		Uri: uri,
	}, nil
}

func (f *FileUsecase) DownloadForRange(ctx context.Context, in *api.DownloadRequest, start int64, end int64) ([]byte, int64, error) {
	if start > end {
		err := errors.New(errors.ErrInvalidRange, int32(api.FileSvrStatus_FILE_INVAILDATE_RANGE_ERR), codes.InvalidArgument, "invalid_range")
		return nil, 0, err

	}
	fileDir := filepath.Join(f.config.GetUploadDir(), in.Key)
	file, err := os.Open(fileDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.New(err, int32(common.Code_NOT_FOUND), codes.NotFound, "file_not_found")
			return nil, 0, err
		}
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "open_file")
		return nil, 0, err

	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "file_info")
		return nil, 0, err

	}
	var buf []byte
	if end > 0 {
		buf = make([]byte, end-start)
	} else {
		buf = make([]byte, fileInfo.Size()-start)
	}
	log.Printf("start:%d,end:%d,bufsize:%d", start, end, len(buf))
	_, err = file.ReadAt(buf, start)
	if err != nil && err != io.EOF {
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
		return nil, 0, err
	}
	return buf, fileInfo.Size(), nil

}
func (f *FileUsecase) Metadata(ctx context.Context, in *api.FileMetadataRequest) (*api.FileMetadataResponse, error) {
	subDir, filename, err := f.spiltKey(in.Key)
	if err != nil {
		return nil, err
	}
	saveDir := filepath.Join(f.config.GetUploadDir(), subDir)
	filePath := filepath.Join(saveDir, filename)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.New(err, int32(common.Code_NOT_FOUND), codes.NotFound, "file_not_found")
			return nil, err
		}
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "open_file")
		return nil, err

	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "file_info")
		return nil, err

	}
	hasher := sha256.New()

	// 以流式传输的方式将文件内容写入哈希对象
	if _, err := io.Copy(hasher, file); err != nil {
		panic(err)
	}

	// 计算最终的哈希值
	sha256Sum := hasher.Sum(nil)
	buf := make([]byte, 512)
	_, _ = file.Read(buf)
	sha256 := fmt.Sprintf("%x", sha256Sum)
	contentType := "application/octet-stream"
	if val := http.DetectContentType(buf); val != "" {
		contentType = val
	}
	return &api.FileMetadataResponse{
		Size:        fileInfo.Size(),
		ModifyTime:  fileInfo.ModTime().UnixMilli(),
		ContentType: contentType,
		Sha256:      sha256,
		Name:        filename,
		Etag:        fmt.Sprintf("%d-%s", fileInfo.Size(), sha256),
	}, nil
}
func (f *FileUsecase) Download(ctx context.Context, in *api.DownloadRequest) ([]byte, error) {
	if in.Key == "" {
		return nil, errors.New(errors.ErrInvalidFileKey, int32(api.FileSvrStatus_FILE_INVAILDATE_KEY_ERR), codes.InvalidArgument, "invalid_key")

	}
	fileDir := filepath.Join(f.config.GetUploadDir(), in.Key)

	file, err := os.Open(fileDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.New(err, int32(common.Code_NOT_FOUND), codes.NotFound, "file_not_found")
			return nil, err
		}
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "open_file")
		return nil, err

	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "file_info")
		return nil, err

	}
	buf := make([]byte, fileInfo.Size())
	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
		return nil, err
	}

	return buf, nil

}

func (f *FileUsecase) Delete(ctx context.Context, in *api.DeleteRequest) (*api.DeleteResponse, error) {
	subDir, filename, err := f.spiltKey(in.Key)
	if err != nil {
		return nil, err
	}
	saveDir := filepath.Join(f.config.GetUploadDir(), subDir)
	filePath := filepath.Join(saveDir, filename)
	err = os.Remove(filePath)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "remove_file")

	}
	keyParts := f.getPersistenceKeyParts(in.Key)
	err = os.RemoveAll(keyParts)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "remove_parts_dir")
	}
	return &api.DeleteResponse{}, nil
}
func (f *FileUsecase) AddFile(ctx context.Context, files []*common.Files) error {
	return f.repo.UpdateFile(ctx, files)
}
