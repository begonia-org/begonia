package file

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	api "github.com/begonia-org/go-sdk/api/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
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
	return "", key, nil
}
func (f *FileUsecase) InitiateUploadFile(ctx context.Context, in *api.InitiateMultipartUploadRequest) (*api.InitiateMultipartUploadResponse, error) {
	if in.Key == "" || strings.HasPrefix(in.Key, "/") {
		return nil, errors.New(errors.ErrInvalidFileKey, int32(api.FileSvrStatus_FILE_INVAILDATE_KEY_ERR), codes.InvalidArgument, "invalid_key")
	}
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

func getSHA256(data []byte) string {
	hash := sha256.Sum256([]byte(data))

	// 将哈希值格式化为十六进制字符串
	hashHex := fmt.Sprintf("%x", hash)
	return hashHex
}

// newVersionFile 检查指定目录的仓库状态，适当时初始化仓库，并提交文件
func (f *FileUsecase) commitFile(dir string, filename string, authorId string, authorEmail string) (commitId string, err error) {
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
			err = w.Reset(&git.ResetOptions{Mode: git.HardReset})
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
	if err != nil {
		return "", err
	}

	// 打印新提交的ID
	obj, err := repo.CommitObject(commit)
	if err != nil && err != git.ErrEmptyCommit {
		return "", err
	}
	// 空提交处理
	if err == git.ErrEmptyCommit {
		headRef, err := repo.Head()
		if err != nil {
			return "", err
		}
		return headRef.Hash().String(), nil
	}

	return obj.ID().String(), nil
}

// getSaveDir Get the save directory of the file
//
// The save directory is the directory where the file is saved.
func (f *FileUsecase) getSaveDir(key string) (string, error) {

	saveDir := filepath.Join(f.config.GetUploadDir(), filepath.Dir(key))

	return saveDir, nil

}

// checkIn checks the key and authorId.
//
// If the key is empty or starts with '/', it returns an error.
// The key is not allow start with '/'.
func (f *FileUsecase) checkIn(key string, authorId string) (string, error) {
	if key == "" || strings.HasPrefix(key, "/") {
		return "", errors.New(errors.ErrInvalidFileKey, int32(api.FileSvrStatus_FILE_INVAILDATE_KEY_ERR), codes.InvalidArgument, "invalid_key")
	}
	if authorId == "" {
		return "", errors.New(errors.ErrIdentityMissing, int32(api.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")
	}
	if !strings.HasPrefix(key, authorId) {
		key = authorId + "/" + key
	}
	return key, nil
}

// Upload uploads a file.
//
// The file is saved in the directory specified by the key.
//
// The authorId is used to determine the directory which is as user's home dir where the file is saved.
func (f *FileUsecase) Upload(ctx context.Context, in *api.UploadFileRequest, authorId string) (*api.UploadFileResponse, error) {

	key, err := f.checkIn(in.Key, authorId)
	if err != nil {
		return nil, err
	}
	in.Key = key
	// var err error
	saveDir, err := f.getSaveDir(in.Key)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_save_dir")
	}
	filename := filepath.Base(in.Key)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_upload_dir")
	}
	filePath := filepath.Join(saveDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_file")
	}

	defer file.Close()
	defer func() {
		if err != nil {
			os.Remove(filePath)
		}
	}()
	_, err = file.Write(in.Content)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "write_file")
	}
	uri, err := f.getUri(filePath, authorId)
	if err != nil {
		return nil, err
	}
	sha256Hash := getSHA256(in.Content)
	if sha256Hash != in.Sha256 {
		os.Remove(filePath)
		err = errors.New(errors.ErrSHA256NotMatch, int32(api.FileSvrStatus_FILE_SHA256_NOT_MATCH_ERR), codes.InvalidArgument, "sha256_not_match")
		return nil, err

	}
	commitId := ""
	if in.UseVersion {
		commitId, err = f.commitFile(saveDir, filename, authorId, "fs@begonia.com")
		if err != nil {
			err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "commit_file")
			return nil, err

		}
	}

	return &api.UploadFileResponse{
		Uri:     uri,
		Version: commitId,
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
	uri, err := f.getUri(filePath, "")
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
func (f *FileUsecase) getPersistenceKeyParts(key string, authorId string) string {
	if strings.Contains(key, ".") {
		key = key[:strings.LastIndex(key, ".")]

	}
	return filepath.Join(f.config.GetUploadDir(), "parts", authorId, key)
}
func (f *FileUsecase) getUri(filePath string, authorId string) (string, error) {
	uploadRootDir := filepath.Join(f.config.GetUploadDir(), authorId)
	// if useVersion {
	// 	uploadRootDir = filepath.Join(f.config.GetUploadDir(), "versions", authorId)
	// }
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
func (f *FileUsecase) CompleteMultipartUploadFile(ctx context.Context, in *api.CompleteMultipartUploadRequest, authorId string) (*api.CompleteMultipartUploadResponse, error) {
	key, err := f.checkIn(in.Key, authorId)
	if err != nil {
		return nil, err
	}
	in.Key = key
	partsDir := f.getPartsDir(in.UploadId)
	if !pathExists(partsDir) {
		err := errors.New(errors.ErrUploadIdNotFound, int32(api.FileSvrStatus_FILE_NOT_FOUND_UPLOADID_ERR), codes.NotFound, "upload_id_not_found")
		return nil, err

	}
	files, err := f.getSortedFiles(partsDir)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_sorted_files")

	}

	saveDir, err := f.getSaveDir(in.Key)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_save_dir")

	}
	filename := filepath.Base(in.Key)
	filePath := filepath.Join(saveDir, filename)
	// merge files to uploadDir/key
	err = f.mergeFiles(files, filePath)
	if err != nil {
		return nil, errors.New(fmt.Errorf("merge file error"), int32(common.Code_INTERNAL_ERROR), codes.Internal, "merge_files")
	}
	// the parts file has been merged, remove the parts dir to uploadDir/parts/key
	keyParts := f.getPersistenceKeyParts(in.Key, authorId)
	if err = os.MkdirAll(keyParts, 0755); err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "create_parts_dir")
	}
	err = f.mvDir(partsDir, keyParts)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "mv_dir")

	}
	uri, err := f.getUri(filePath, authorId)
	if err != nil {
		return nil, err

	}
	commit := ""
	if in.UseVersion {
		commit, err = f.commitFile(saveDir, filename, authorId, "begonia@begonia.com")
		if err != nil {
			return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "commit_file")
		}
	}

	os.RemoveAll(filepath.Join(f.config.GetUploadDir(), in.UploadId))
	return &api.CompleteMultipartUploadResponse{
		Uri:     uri,
		Version: commit,
	}, nil
}

func (f *FileUsecase) DownloadForRange(ctx context.Context, in *api.DownloadRequest, start int64, end int64, authorId string) ([]byte, int64, error) {
	key, err := f.checkIn(in.Key, authorId)
	if err != nil {
		return nil, 0, err
	}
	in.Key = key
	if start > end {
		err := errors.New(errors.ErrInvalidRange, int32(api.FileSvrStatus_FILE_INVAILDATE_RANGE_ERR), codes.InvalidArgument, "invalid_range")
		return nil, 0, err

	}

	file, err := f.getReader(in.Key, in.Version)
	if err == git.ErrRepositoryNotExists || os.IsNotExist(err) {
		return nil, 0, errors.New(err, int32(common.Code_NOT_FOUND), codes.NotFound, "file_not_found")
	}
	if err != nil {
		return nil, 0, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "open_file")
	}
	defer file.Close()

	var buf []byte
	if end > 0 {
		buf = make([]byte, end-start)
	} else {
		buf = make([]byte, file.Size()-start)
	}
	// log.Printf("start:%d,end:%d,bufsize:%d", start, end, len(buf))
	_, err = file.ReadAt(buf, start)
	if err != nil && err != io.EOF {
		err = errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
		return nil, 0, err
	}
	return buf, file.Size(), nil

}
func (f *FileUsecase) Metadata(ctx context.Context, in *api.FileMetadataRequest, authorId string) (*api.FileMetadataResponse, error) {
	key, err := f.checkIn(in.Key, authorId)
	if err != nil {
		return nil, err
	}
	in.Key = key
	file, err := f.getReader(in.Key, in.Version)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "open_file")

	}
	hasher := sha256.New()
	reader, err := file.Reader()
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
	}
	defer reader.Close()
	defer file.Close()
	// 以流式传输的方式将文件内容写入哈希对象
	if _, err := io.Copy(hasher, reader); err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "hash_file")
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

	return &api.FileMetadataResponse{
		Size:        file.Size(),
		ModifyTime:  file.ModifyTime(),
		ContentType: contentType,
		Sha256:      sha256,
		Name:        filepath.Base(in.Key),
		Etag:        fmt.Sprintf("%d-%s", file.Size(), sha256),
	}, nil
}

// getReader obtains a file reader.
//
// If the version is not empty, the file reader will be a version file reader.
func (f *FileUsecase) getReader(key string, version string) (FileReader, error) {

	dir, err := f.getSaveDir(key)

	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(dir, filepath.Base(key))
	var fileReader FileReader
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

func (f *FileUsecase) Download(ctx context.Context, in *api.DownloadRequest, authorId string) ([]byte, error) {
	// if in.Key == "" {
	// 	return nil, errors.New(errors.ErrInvalidFileKey, int32(api.FileSvrStatus_FILE_INVAILDATE_KEY_ERR), codes.InvalidArgument, "invalid_key")

	// }
	key, err := f.checkIn(in.Key, authorId)
	if err != nil {
		return nil, err
	}
	in.Key = key
	// fileDir := filepath.Join(f.config.GetUploadDir(), in.Key)
	file, err := f.getReader(in.Key, in.Version)
	if err == git.ErrRepositoryNotExists {
		return nil, errors.New(err, int32(common.Code_NOT_FOUND), codes.NotFound, "file_not_found")
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.New(os.ErrNotExist, int32(common.Code_INTERNAL_ERROR), codes.Internal, "open_file")
	}
	if os.IsNotExist(err) {
		return nil, errors.New(err, int32(common.Code_NOT_FOUND), codes.NotFound, "file_not_found")
	}
	buf := make([]byte, file.Size())
	reader, err := file.Reader()
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
	}
	defer reader.Close()
	defer file.Close()
	_, err = reader.Read(buf)
	if err != nil && err != io.EOF {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_file")
	}
	return buf, nil

}

func (f *FileUsecase) Delete(ctx context.Context, in *api.DeleteRequest, authorId string) (*api.DeleteResponse, error) {
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
	keyParts := f.getPersistenceKeyParts(in.Key, authorId)
	err = os.RemoveAll(keyParts)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "remove_parts_dir")
	}
	return &api.DeleteResponse{}, nil
}
func (f *FileUsecase) AddFile(ctx context.Context, files []*common.Files) error {
	return f.repo.UpdateFile(ctx, files)
}
