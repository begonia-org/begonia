package file

import (
	"context"
	"crypto/sha256"
	goErr "errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/pkg"
	"github.com/begonia-org/begonia/internal/pkg/config"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/file/v1"
	user "github.com/begonia-org/go-sdk/api/user/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spark-lence/tiga"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FileRepo interface {
	// mysql
	UpsertFile(ctx context.Context, file *api.Files) error
	DelFile(ctx context.Context, engine, bucket, key string) error
	UpsertBucket(ctx context.Context, bucket *api.Buckets) error
	DelBucket(ctx context.Context, bucketId string) error
	List(ctx context.Context, page, pageSize int32, bucket, engine, owner string) ([]*api.Files, error)
}
type FileUsecase interface {
	Upload(ctx context.Context, in *api.UploadFileRequest, authorId string) (*api.UploadFileResponse, error)
	InitiateUploadFile(ctx context.Context, in *api.InitiateMultipartUploadRequest) (*api.InitiateMultipartUploadResponse, error)
	UploadMultipartFileFile(ctx context.Context, in *api.UploadMultipartFileRequest) (*api.UploadMultipartFileResponse, error)
	AbortMultipartUpload(ctx context.Context, in *api.AbortMultipartUploadRequest) (*api.AbortMultipartUploadResponse, error)
	CompleteMultipartUploadFile(ctx context.Context, in *api.CompleteMultipartUploadRequest, authorId string) (*api.CompleteMultipartUploadResponse, error)
	DownloadForRange(ctx context.Context, in *api.DownloadRequest, start int64, end int64, authorId string) ([]byte, int64, error)
	Metadata(ctx context.Context, in *api.FileMetadataRequest, authorId string) (*api.FileMetadataResponse, error)
	Version(ctx context.Context, bucket, key, authorId string) (string, error)
	Download(ctx context.Context, in *api.DownloadRequest, authorId string) ([]byte, error)
	Delete(ctx context.Context, in *api.DeleteRequest, authorId string) (*api.DeleteResponse, error)
	MakeBucket(ctx context.Context, in *api.MakeBucketRequest) (*api.MakeBucketResponse, error)
	List(ctx context.Context, in *api.ListFilesRequest, authorId string) ([]*api.Files, error)
}
type FileUsecaseImpl struct {
	repo      FileRepo
	config    *config.Config
	snowflake *tiga.Snowflake
}

func newMinioClient(cfg *config.Config) *minio.Client {
	engines, err := cfg.GetFileEngines()
	if err != nil {
		panic(err)

	}
	for _, engine := range engines {
		if engine.Name == api.FileEngine_FILE_ENGINE_MINIO.String() {
			endpoint := engine.Endpoint
			creds := credentials.NewStaticV4(engine.GetAccessKey(), engine.SecretKey, "")
			minioClient, err := minio.New(endpoint, &minio.Options{
				Creds:  creds,
				Secure: false,
			})
			if err != nil {
				panic(err)
			}
			return minioClient
		}
	}
	return nil

}
func NewFileUsecase(cfg *config.Config, repo FileRepo) map[string]FileUsecase {
	engines := make(map[string]FileUsecase)
	if minioClient := newMinioClient(cfg); minioClient != nil {
		engines[api.FileEngine_FILE_ENGINE_MINIO.String()] = NewMinioUseCase(minioClient, NewLocalFileUsecase(cfg, repo))
	}
	engines[api.FileEngine_FILE_ENGINE_LOCAL.String()] = NewLocalFileUsecase(cfg, repo)
	return engines
}
func NewLocalFileUsecase(config *config.Config, repo FileRepo) *FileUsecaseImpl {
	snk, _ := tiga.NewSnowflake(1)
	return &FileUsecaseImpl{config: config, snowflake: snk, repo: repo}
}
func (f *FileUsecaseImpl) getPartsDir(key string) string {
	return filepath.Join(f.config.GetUploadDir(), key, "parts")
}
func (f *FileUsecaseImpl) InitiateUploadFile(ctx context.Context, in *api.InitiateMultipartUploadRequest) (*api.InitiateMultipartUploadResponse, error) {
	if in.Key == "" || strings.HasPrefix(in.Key, "/") {
		return nil, gosdk.NewError(pkg.ErrInvalidFileKey, int32(api.FileSvrStatus_FILE_INVALIDATE_KEY_ERR), codes.InvalidArgument, "invalid_key")
	}
	uploadId := f.snowflake.GenerateIDString()
	saveDir := f.getPartsDir(uploadId)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		err = gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_upload_dir")
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

func getSHA256(data []byte) string {
	hash := sha256.Sum256([]byte(data))

	// 将哈希值格式化为十六进制字符串
	hashHex := fmt.Sprintf("%x", hash)
	return hashHex
}

// newVersionFile 检查指定目录的仓库状态，适当时初始化仓库，并提交文件
func (f *FileUsecaseImpl) commitFile(dir string, filename string, authorId string, authorEmail string) (commitId string, err error) {
	repo, err := git.PlainInit(dir, false)
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		return "", err
	}

	// 如果仓库已存在，则打开它
	if err == git.ErrRepositoryAlreadyExists {
		repo, err = git.PlainOpen(dir)
		if err != nil {
			return "", err
		}
	}

	// 设置作者信息
	author := &object.Signature{
		Name:  authorId,
		Email: authorEmail,
		When:  time.Now(),
	}

	// 工作树
	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	defer func() {
		if p := recover(); p != nil {
			err = p.(error)
		}
		if err != nil {
			_ = w.Reset(&git.ResetOptions{Mode: git.HardReset})
		}

	}()
	// 添加文件到暂存区
	_, err = w.Add(filename)
	if err != nil {
		return "", err
	}

	// 创建提交
	commit, err := w.Commit(fmt.Sprintf("Add %s", filename), &git.CommitOptions{
		Author: author,
	})
	if err != nil && !goErr.Is(err, git.ErrEmptyCommit) {
		return "", err
	}
	// 空提交处理
	if goErr.Is(err, git.ErrEmptyCommit) {

		headRef, err := repo.Head()
		if err != nil || headRef.Hash().IsZero() {
			return "", fmt.Errorf("get head ref error:%w or head ref is nil", err)
		}
		return headRef.Hash().String(), nil
	}
	// 打印新提交的ID
	obj, err := repo.CommitObject(commit)

	if err != nil {
		return "", err
	}

	return obj.ID().String(), nil
}

// getSaveDir Get the save directory of the file
//
// The save directory is the directory where the file is saved.
func (f *FileUsecaseImpl) getSaveDir(bucket, key string) string {

	saveDir := filepath.Join(f.config.GetUploadDir(), bucket, filepath.Dir(key))

	return saveDir

}

// checkIn checks the key and authorId.
//
// If the key is empty or starts with '/', it returns an error.
// The key is not allow start with '/'.
func (f *FileUsecaseImpl) checkIn(key string) (string, error) {
	if key == "" || strings.HasPrefix(key, "/") {
		return "", gosdk.NewError(pkg.ErrInvalidFileKey, int32(api.FileSvrStatus_FILE_INVALIDATE_KEY_ERR), codes.InvalidArgument, "invalid_key")
	}
	return key, nil
}
func (f *FileUsecaseImpl) MakeBucket(ctx context.Context, in *api.MakeBucketRequest) (*api.MakeBucketResponse, error) {
	saveDir := filepath.Join(f.config.GetUploadDir(), in.Bucket)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "make_bucket")
	}
	return &api.MakeBucketResponse{}, nil
}
func (f *FileUsecaseImpl) checkBucket(bucket string) bool {
	saveDir := filepath.Join(f.config.GetUploadDir(), bucket)
	return pathExists(saveDir)
}

// Upload uploads a file.
//
// The file is saved in the directory specified by the key.
//
// The authorId is used to determine the directory which is as user's home dir where the file is saved.
func (f *FileUsecaseImpl) Upload(ctx context.Context, in *api.UploadFileRequest, authorId string) (*api.UploadFileResponse, error) {
	if authorId == "" {
		return nil, gosdk.NewError(pkg.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")
	}

	key, err := f.checkIn(in.Key)
	if err != nil {
		return nil, err
	}
	in.Key = filepath.Join(authorId, key)
	// log.Printf("key:%s", in.Key)
	filename := filepath.Base(in.Key)
	if ok := f.checkBucket(in.Bucket); in.Bucket == "" || !ok {
		return nil, gosdk.NewError(pkg.ErrBucketNotFound, int32(api.FileSvrStatus_FILE_INVALIDATE_BUCKET_ERR), codes.NotFound, "bucket_not_found")

	}
	saveDir := f.getSaveDir(in.Bucket, in.Key)

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_upload_dir")
	}
	filePath := filepath.Join(saveDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_file")
	}
	defer file.Close()
	defer func() {
		if err != nil {
			os.Remove(filePath)
		}
	}()
	_, err = file.Write(in.Content)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "write_file")
	}
	uri, err := f.getUri(in.Bucket, filePath)
	if err != nil {
		return nil, err
	}
	sha256Hash := getSHA256(in.Content)
	if sha256Hash != in.Sha256 {
		os.Remove(filePath)
		err = gosdk.NewError(pkg.ErrSHA256NotMatch, int32(api.FileSvrStatus_FILE_SHA256_NOT_MATCH_ERR), codes.InvalidArgument, "sha256_not_match")
		return nil, err

	}
	commitId := ""
	if in.UseVersion {
		commitId, err = f.commitFile(saveDir, filename, authorId, "fs@begonia.com")
		if err != nil {
			err = gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "commit_file")
			return nil, err

		}
	}
	err = f.repo.UpsertFile(ctx, &api.Files{
		Uid:       f.snowflake.GenerateIDString(),
		Engine:    api.FileEngine_name[int32(api.FileEngine_FILE_ENGINE_LOCAL)],
		Bucket:    in.Bucket,
		Key:       uri,
		IsDeleted: false,
		Owner:     authorId,
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	})
	if err != nil {
		return nil, err
	}
	return &api.UploadFileResponse{
		Uri:     uri,
		Version: commitId,
	}, err

}
func (f *FileUsecaseImpl) UploadMultipartFileFile(ctx context.Context, in *api.UploadMultipartFileRequest) (*api.UploadMultipartFileResponse, error) {

	if in.UploadId == "" {
		return nil, gosdk.NewError(pkg.ErrUploadIdMissing, int32(api.FileSvrStatus_FILE_UPLOADID_MISSING_ERR), codes.InvalidArgument, "upload_id_not_found")
	}
	if in.PartNumber <= 0 {
		return nil, gosdk.NewError(pkg.ErrPartNumberMissing, int32(api.FileSvrStatus_FILE_PARTNUMBER_MISSING_ERR), codes.InvalidArgument, "part_number_not_found")

	}

	uploadId := in.UploadId
	// get upload dir by uploadId
	saveDir := f.getPartsDir(uploadId)
	if !pathExists(saveDir) {
		err := gosdk.NewError(pkg.ErrUploadNotInitiate, int32(api.FileSvrStatus_FILE_UPLOAD_NOT_INITIATE_ERR), codes.NotFound, "upload_dir_not_found")
		return nil, err
	}

	filePath := filepath.Join(saveDir, fmt.Sprintf("%08d.part", in.PartNumber))

	file, err := os.Create(filePath)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_file")
	}
	defer file.Close()
	_, err = file.Write(in.Content)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "write_file")

	}
	sha256Hash := getSHA256(in.Content)
	if sha256Hash != in.Sha256 {
		os.Remove(filePath)
		err := gosdk.NewError(pkg.ErrSHA256NotMatch, int32(api.FileSvrStatus_FILE_SHA256_NOT_MATCH_ERR), codes.InvalidArgument, "sha256_not_match")
		return nil, err

	}
	// uri, err := f.getUri(in.,filePath)
	// if err != nil {
	// 	return nil, err
	// }
	return &api.UploadMultipartFileResponse{
		// Uri: uri,
	}, nil
}
func (f *FileUsecaseImpl) getSortedFiles(dirPath string) ([]string, error) {
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
	sort.Strings(files)
	return files, nil
}
func (f *FileUsecaseImpl) mergeFiles(files []string, outputFile string) error {
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
func (f *FileUsecaseImpl) mvDir(src, dst string) error {
	newPath := filepath.Join(dst, filepath.Base(src))
	if _, err := os.Stat(newPath); err == nil {
		// 目标路径存在，尝试删除
		if err := os.RemoveAll(newPath); err != nil {
			return err
		}
	}

	return os.Rename(src, newPath)
}
func (f *FileUsecaseImpl) getPersistenceKeyParts(key string) string {
	if strings.Contains(key, ".") {
		key = key[:strings.LastIndex(key, ".")]

	}
	return filepath.Join(f.config.GetUploadDir(), "parts", key)
}
func (f *FileUsecaseImpl) getUri(bucket, filePath string) (string, error) {
	uploadRootDir := f.config.GetUploadDir()
	// log.Printf("uploadRootDir:%s,filePath:%s", uploadRootDir, filePath)
	uri, err := filepath.Rel(filepath.Join(uploadRootDir, bucket), filePath)
	if err != nil {
		return "", gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_file_uri")
	}
	return uri, nil

}
func (f *FileUsecaseImpl) AbortMultipartUpload(ctx context.Context, in *api.AbortMultipartUploadRequest) (*api.AbortMultipartUploadResponse, error) {
	partsDir := f.getPartsDir(in.UploadId)
	if !pathExists(partsDir) {
		err := gosdk.NewError(pkg.ErrUploadIdNotFound, int32(api.FileSvrStatus_FILE_NOT_FOUND_UPLOADID_ERR), codes.NotFound, "upload_id_not_found")
		return nil, err

	}
	err := os.RemoveAll(partsDir)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "remove_parts_dir")
	}
	return &api.AbortMultipartUploadResponse{}, nil
}
func (f *FileUsecaseImpl) CompleteMultipartUploadFile(ctx context.Context, in *api.CompleteMultipartUploadRequest, authorId string) (*api.CompleteMultipartUploadResponse, error) {
	if authorId == "" {
		return nil, gosdk.NewError(pkg.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")
	}
	key, err := f.checkIn(in.Key)
	if err != nil {
		return nil, err
	}
	if ok := f.checkBucket(in.Bucket); in.Bucket == "" || !ok {
		return nil, gosdk.NewError(pkg.ErrBucketNotFound, int32(api.FileSvrStatus_FILE_INVALIDATE_BUCKET_ERR), codes.NotFound, "bucket_not_found")

	}
	in.Key = filepath.Join(authorId, key)
	partsDir := f.getPartsDir(in.UploadId)
	if !pathExists(partsDir) {
		err := gosdk.NewError(fmt.Errorf("%s:%s", in.UploadId, pkg.ErrUploadIdNotFound.Error()), int32(api.FileSvrStatus_FILE_NOT_FOUND_UPLOADID_ERR), codes.NotFound, "upload_id_not_found")
		return nil, err

	}
	files, err := f.getSortedFiles(partsDir)
	if err != nil {

		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_sorted_files")

	}

	saveDir := f.getSaveDir(in.Bucket, in.Key)
	filename := filepath.Base(in.Key)
	filePath := filepath.Join(saveDir, filename)
	err = f.mergeFiles(files, filePath)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("merge file error:%w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "merge_files")
	}
	// the parts file has been merged, remove the parts dir to uploadDir/parts/key
	keyParts := f.getPersistenceKeyParts(in.Key)
	if err = os.MkdirAll(keyParts, 0755); err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_parts_dir")
	}
	err = f.mvDir(partsDir, keyParts)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "mv_dir")

	}
	uri, err := f.getUri(in.Bucket, filePath)
	if err != nil {
		return nil, err

	}
	commit := ""
	if in.UseVersion {
		commit, err = f.commitFile(saveDir, filename, authorId, "begonia@begonia.com")
		if err != nil {
			return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "commit_file")
		}
	}

	err = f.repo.UpsertFile(ctx, &api.Files{
		Uid:       f.snowflake.GenerateIDString(),
		Engine:    in.Engine,
		Bucket:    in.Bucket,
		Key:       uri,
		IsDeleted: false,
		Owner:     authorId,
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	})
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "upsert_file")
	}
	os.RemoveAll(filepath.Join(f.config.GetUploadDir(), in.UploadId))

	return &api.CompleteMultipartUploadResponse{
		Uri:     uri,
		Version: commit,
	}, err
}

func (f *FileUsecaseImpl) DownloadForRange(ctx context.Context, in *api.DownloadRequest, start int64, end int64, authorId string) ([]byte, int64, error) {
	key, err := f.checkIn(in.Key)
	if err != nil {
		return nil, 0, err
	}
	in.Key = key
	if start > end && end > 0 {
		err := gosdk.NewError(fmt.Errorf("%w:start=%d,end=%d", pkg.ErrInvalidRange, start, end), int32(api.FileSvrStatus_FILE_INVALIDATE_RANGE_ERR), codes.InvalidArgument, "invalid_range")
		return nil, 0, err

	}

	file, err := f.getReader(in.Bucket, in.Key, in.Version)
	if err != nil {
		code, grcpCode := f.checkStatusCode(err)
		return nil, 0, gosdk.NewError(err, code, grcpCode, "open_file")
	}
	defer file.Close()

	var buf []byte
	if end > 0 {
		buf = make([]byte, end-start+1)
	} else {
		buf = make([]byte, file.Size()-start)
	}
	_, err = file.ReadAt(buf, start)
	if err != nil && err != io.EOF {
		err = gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
		return nil, 0, err
	}
	return buf, file.Size(), nil

}
func (f *FileUsecaseImpl) Metadata(ctx context.Context, in *api.FileMetadataRequest, authorId string) (*api.FileMetadataResponse, error) {
	key, err := f.checkIn(in.Key)
	originKey := in.Key
	if err != nil {
		return nil, err
	}
	in.Key = key
	file, err := f.getReader(in.Bucket, in.Key, in.Version)
	if err != nil {
		code, grpcCode := f.checkStatusCode(err)
		return nil, gosdk.NewError(err, code, grpcCode, "open_file")

	}
	hasher := sha256.New()
	reader, err := file.Reader()
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
	}
	defer reader.Close()
	defer file.Close()
	// 以流式传输的方式将文件内容写入哈希对象
	if _, err := io.Copy(hasher, reader); err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "hash_file")
	}

	// 计算最终的哈希值
	sha256Sum := hasher.Sum(nil)
	buf := make([]byte, 512)
	_, _ = reader.Read(buf)
	sha256 := fmt.Sprintf("%x", sha256Sum)
	contentType := "application/octet-stream"
	if val := http.DetectContentType(buf); val != "" {
		contentType = val
	}
	version := ""
	if versionReader, ok := file.(FileVersionReader); ok {
		version = versionReader.Version()
	} else {
		version, _ = f.Version(ctx, in.Bucket, originKey, authorId)

	}
	return &api.FileMetadataResponse{
		Size:        file.Size(),
		ModifyTime:  file.ModifyTime(),
		ContentType: contentType,
		Sha256:      sha256,
		Name:        filepath.Base(in.Key),
		Etag:        fmt.Sprintf("%d-%s", file.Size(), sha256),
		Version:     version,
	}, nil
}

// getReader obtains a file reader.
//
// If the version is not empty, the file reader will be a version file reader.
func (f *FileUsecaseImpl) getReader(bucket, key string, version string) (FileReader, error) {

	dir := f.getSaveDir(bucket, key)

	filePath := filepath.Join(dir, filepath.Base(key))
	var fileReader FileReader
	var err error
	if version != "" {
		fileReader, err = NewFileVersionReader(filePath, version)
		if err != nil {
			return nil, err
		}
	}
	if fileReader == nil {
		fileReader, err = NewFileReader(filePath)
		if err != nil {
			return nil, err
		}
	}
	return fileReader, nil

}
func (f *FileUsecaseImpl) Version(ctx context.Context, bucket, key, authorId string) (string, error) {
	key, err := f.checkIn(key)
	if err != nil {
		return "", err
	}
	file, err := f.getReader(bucket, key, "latest")
	if err != nil {
		code, grpcCode := f.checkStatusCode(err)
		return "", gosdk.NewError(err, code, grpcCode, "open_file")
	}
	defer file.Close()
	return file.(FileVersionReader).Version(), nil
}
func (f *FileUsecaseImpl) checkStatusCode(err error) (int32, codes.Code) {

	switch err {
	case git.ErrRepositoryNotExists, os.ErrNotExist:
		return int32(common.Code_NOT_FOUND), codes.NotFound
	default:
		return int32(common.Code_INTERNAL_ERROR), codes.Internal
	}
}
func (f *FileUsecaseImpl) Download(ctx context.Context, in *api.DownloadRequest, authorId string) ([]byte, error) {

	key, err := f.checkIn(in.Key)
	if err != nil {
		return nil, err
	}
	in.Key = key
	file, err := f.getReader(in.Bucket, in.Key, in.Version)
	if err != nil {
		code, httpCode := f.checkStatusCode(err)
		return nil, gosdk.NewError(err, code, httpCode, "open_file")
	}
	buf := make([]byte, file.Size())
	reader, err := file.Reader()
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
	}
	defer reader.Close()
	defer file.Close()
	_, err = reader.Read(buf)
	if err != nil && err != io.EOF {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
	}
	return buf, nil

}
func (f *FileUsecaseImpl) Delete(ctx context.Context, in *api.DeleteRequest, authorId string) (*api.DeleteResponse, error) {
	key, err := f.checkIn(in.Key)
	if err != nil {
		return nil, err
	}
	in.Key = key
	file, err := f.getReader(in.Bucket, in.Key, "")
	if err != nil && !os.IsNotExist(err) {
		// log.Printf("err:%v", err)
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "remove_file")

	}
	if file != nil {
		defer file.Close()
		os.Remove(file.Name())
	}
	versionFile, err := f.getReader(in.Bucket, in.Key, "latest")
	if err != nil {
		// log.Printf("version err:%v", err)
		code, rpcCode := f.checkStatusCode(err)
		return nil, gosdk.NewError(err, code, rpcCode, "remove_file")
	}
	defer versionFile.Close()
	os.Remove(versionFile.Name())

	keyParts := f.getPersistenceKeyParts(in.Key)
	err = os.RemoveAll(keyParts)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "remove_parts_dir")
	}
	err = f.repo.DelFile(ctx, in.Bucket, in.Key, in.Key)
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "remove_file")

	}
	return &api.DeleteResponse{}, err
}

func (f *FileUsecaseImpl) List(ctx context.Context, in *api.ListFilesRequest, authorId string) ([]*api.Files, error) {
	files, err := f.repo.List(ctx, in.Page, int32(in.PageSize), in.Bucket, in.Engine, authorId)
	if err != nil {
		return nil, err
	}
	return files, nil
}
