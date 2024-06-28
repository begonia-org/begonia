package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/begonia-org/begonia/internal/pkg"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/file/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MinioUseCase struct {
	minioClient *minio.Client
	localFile   *FileUsecaseImpl
}

func NewMinioUseCase(minioClient *minio.Client, localFile *FileUsecaseImpl) FileUsecase {
	return &MinioUseCase{minioClient: minioClient, localFile: localFile}
}
func (m *MinioUseCase) MakeBucket(ctx context.Context, in *api.MakeBucketRequest, _ string) (*api.MakeBucketResponse, error) {
	err := m.minioClient.MakeBucket(ctx, in.Bucket, minio.MakeBucketOptions{Region: in.Region, ObjectLocking: in.ObjectLocking})
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to make bucket: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "make_bucket", gosdk.WithClientMessage(err.Error()))
	}
	err = m.minioClient.EnableVersioning(ctx, in.Bucket)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to enable versioning: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "enable_versioning", gosdk.WithClientMessage(err.Error()))

	}

	return &api.MakeBucketResponse{}, nil
}
func (m *MinioUseCase) Upload(ctx context.Context, in *api.UploadFileRequest, authorId string) (*api.UploadFileResponse, error) {
	found, err := m.minioClient.BucketExists(ctx, in.Bucket)
	if err != nil || !found {
		return nil, gosdk.NewError(fmt.Errorf("failed to check bucket: %w or bucket %s not exist", err, in.Bucket), int32(common.Code_INTERNAL_ERROR), codes.Internal, "check_bucket")
	}
	contentLength := len(in.Content)

	in.Engine = api.FileEngine_FILE_ENGINE_MINIO.String()
	hash := sha256.New()
	hash.Write([]byte(in.Content))
	checksum := hash.Sum(nil)

	// 设置对象元数据，包括 SHA-256 校验和
	userMetadata := map[string]string{
		"x-amz-checksum-algorithm": "SHA256",
		"x-amz-checksum-sha256":    base64.StdEncoding.EncodeToString(checksum),
	}
	log.Printf("Put object %s to %s,and its size is %d", in.Key, in.Bucket, contentLength)
	info, err := m.minioClient.PutObject(ctx, in.Bucket, in.Key, bytes.NewReader(in.Content), int64(contentLength), minio.PutObjectOptions{ContentType: in.ContentType, DisableContentSha256: false, UserMetadata: userMetadata})
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to upload object: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "upload_object")

	}
	f := &api.Files{
		Uid:       m.localFile.snowflake.GenerateIDString(),
		Engine:    in.Engine,
		Bucket:    in.Bucket,
		Key:       in.Key,
		IsDeleted: false,
		Owner:     authorId,
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}
	updated, err := m.localFile.repo.UpsertFile(ctx, f)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to upsert file: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "upsert_file")

	}

	uid := f.Uid
	if updated {
		existsObj, err := m.localFile.repo.GetFile(ctx, f.Engine, f.Bucket, f.Key)
		if err != nil {
			return nil, gosdk.NewError(fmt.Errorf("failed to get updated file info:%w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_file")
		}
		if existsObj != nil {
			uid = existsObj.Uid
		}
	}
	// log.Printf("upload object:%s,%s", info.Key, info.ChecksumSHA256)
	return &api.UploadFileResponse{Uri: info.Key, Version: info.VersionID, Uid: uid}, err
}
func (m *MinioUseCase) InitiateUploadFile(ctx context.Context, in *api.InitiateMultipartUploadRequest) (*api.InitiateMultipartUploadResponse, error) {
	in.Engine = api.FileEngine_FILE_ENGINE_MINIO.String()
	return m.localFile.InitiateUploadFile(ctx, in)
}
func (m *MinioUseCase) UploadMultipartFileFile(ctx context.Context, in *api.UploadMultipartFileRequest) (*api.UploadMultipartFileResponse, error) {
	in.Engine = api.FileEngine_FILE_ENGINE_MINIO.String()
	return m.localFile.UploadMultipartFileFile(ctx, in)
}
func (m *MinioUseCase) AbortMultipartUpload(ctx context.Context, in *api.AbortMultipartUploadRequest) (*api.AbortMultipartUploadResponse, error) {
	in.Engine = api.FileEngine_FILE_ENGINE_MINIO.String()
	return m.localFile.AbortMultipartUpload(ctx, in)
}
func (m *MinioUseCase) CompleteMultipartUploadFile(ctx context.Context, in *api.CompleteMultipartUploadRequest, authorId string) (*api.CompleteMultipartUploadResponse, error) {
	// _, _ = m.localFile.MakeBucket(ctx, &api.MakeBucketRequest{Bucket: in.Bucket},authorId)
	originKey := in.Key
	in.Engine = api.FileEngine_FILE_ENGINE_MINIO.String()
	rsp, err := m.localFile.CompleteMultipartUploadFile(ctx, in, authorId)
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(m.localFile.config.GetUploadDir(), in.Bucket, rsp.Uri)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to open file: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "open_file")

	}
	defer file.Close()
	stat, _ := os.Stat(filePath)
	// if in.UseVersion {
	// 	_ = m.minioClient.EnableVersioning(ctx, in.Bucket)
	// }

	userMetadata := map[string]string{
		"x-amz-meta-sha256": in.Sha256,
	}
	info, err := m.minioClient.PutObject(ctx, in.Bucket, originKey, file, stat.Size(), minio.PutObjectOptions{ContentType: in.ContentType, DisableContentSha256: false, PartSize: 1024 * 1024 * 8, ConcurrentStreamParts: true, UserMetadata: userMetadata})
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to upload object: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "upload_object")
	}
	defer os.Remove(filePath)

	return &api.CompleteMultipartUploadResponse{Uri: info.Key, Version: info.VersionID}, nil
}
func (m *MinioUseCase) DownloadForRange(ctx context.Context, in *api.DownloadRequest, start int64, end int64, authorId string) ([]byte, int64, error) {
	object, err := m.minioClient.GetObject(ctx, in.Bucket, in.Key, minio.GetObjectOptions{VersionID: in.Version, Checksum: true})
	if err != nil {
		return nil, 0, gosdk.NewError(fmt.Errorf("failed to get %s %s object: %w", in.Bucket, in.Key, err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_object")
	}
	if start > end && end > 0 {
		err := gosdk.NewError(fmt.Errorf("%w:start=%d,end=%d", pkg.ErrInvalidRange, start, end), int32(api.FileSvrStatus_FILE_INVALIDATE_RANGE_ERR), codes.InvalidArgument, "invalid_range")
		return nil, 0, err

	}
	stat, err := object.Stat()
	if err != nil {
		return nil, 0, gosdk.NewError(fmt.Errorf("failed to get %s/%s object stat: %w", in.Bucket, in.Key, err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_object_stat")

	}
	var buf []byte
	if end > 0 {
		buf = make([]byte, end-start+1)
	} else {
		buf = make([]byte, stat.Size-start)
	}
	_, err = object.ReadAt(buf, start)
	if err != nil && err != io.EOF {
		return nil, 0, gosdk.NewError(fmt.Errorf("failed to read object: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_object")

	}
	return buf, stat.Size, nil

}
func (m *MinioUseCase) Metadata(ctx context.Context, in *api.FileMetadataRequest, authorId string) (*api.FileMetadataResponse, error) {

	opt := minio.StatObjectOptions{VersionID: in.Version, Checksum: true}

	info, err := m.minioClient.StatObject(ctx, in.Bucket, in.Key, opt)
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to get object stat: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_object_stat")
	}

	checksum := info.ChecksumSHA256
	if checksum != "" {
		c, _ := base64.StdEncoding.DecodeString(checksum)
		checksum = hex.EncodeToString(c)
	}
	if checksum == "" {
		checksum = info.UserMetadata["Sha256"]
	}
	fileId := in.FileId
	bucket := in.Bucket

	if fileId == "" {
		// log.Printf("fileId is empty,try to get file info from local file,%s,%s,%s", in.Engine, in.Bucket, in.Key)
		file, err := m.localFile.repo.GetFile(ctx, in.Engine, in.Bucket, in.Key)
		if err != nil {
			return nil, gosdk.NewError(fmt.Errorf("failed to get file: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_file")
		}
		fileId = file.Uid
		bucket = file.Bucket
	}
	// log.Printf("fileId:%s,bucket:%s", fileId, bucket)
	return &api.FileMetadataResponse{Size: info.Size,
		ContentType: info.ContentType,
		Version:     info.VersionID,
		ModifyTime:  int64(info.LastModified.Second()),
		Etag:        info.ETag,
		Name:        info.Key,
		Sha256:      checksum,
		Uid:         fileId,
		Bucket:      bucket,
	}, nil
}
func (m *MinioUseCase) Version(ctx context.Context, bucket, key, authorId string) (string, error) {
	info, err := m.minioClient.StatObject(ctx, bucket, key, minio.StatObjectOptions{Checksum: true})
	if err != nil {
		return "", gosdk.NewError(fmt.Errorf("failed to get object stat: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_object_stat")
	}
	return info.VersionID, nil
}
func (m *MinioUseCase) Download(ctx context.Context, in *api.DownloadRequest, authorId string) ([]byte, error) {
	object, err := m.minioClient.GetObject(ctx, in.Bucket, in.Key, minio.GetObjectOptions{})
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to get object: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_object")
	}
	stat, err := object.Stat()
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to get object stat: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "get_object_stat")

	}
	buf := make([]byte, stat.Size)
	_, err = object.Read(buf)
	if err != nil && err != io.EOF {
		return nil, gosdk.NewError(fmt.Errorf("failed to read object: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "read_object")

	}
	// log.Printf("download object:%s,size:%d,%v", in.Key,len(buf),stat.Size)
	return buf, nil
}
func (m *MinioUseCase) Delete(ctx context.Context, in *api.DeleteRequest, authorId string) (*api.DeleteResponse, error) {
	err := m.minioClient.RemoveObject(ctx, in.Bucket, in.Key, minio.RemoveObjectOptions{VersionID: in.Version})
	if err != nil {
		return nil, gosdk.NewError(fmt.Errorf("failed to remove object: %w", err), int32(common.Code_INTERNAL_ERROR), codes.Internal, "remove_object")
	}
	return &api.DeleteResponse{}, nil
}
func (m *MinioUseCase) List(ctx context.Context, in *api.ListFilesRequest, authorId string) ([]*api.Files, error) {
	in.Engine = api.FileEngine_FILE_ENGINE_MINIO.String()
	return m.localFile.List(ctx, in, authorId)
}
func (m *MinioUseCase) GetFileByID(ctx context.Context, fileId string) (*api.Files, error) {
	return m.localFile.GetFileByID(ctx, fileId)
}
