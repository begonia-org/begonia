package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	gosdk "github.com/begonia-org/go-sdk"
	api "github.com/begonia-org/go-sdk/api/file/v1"
	user "github.com/begonia-org/go-sdk/api/user/v1"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type FileService struct {
	api.UnimplementedFileServiceServer
	biz    *file.FileUsecase
	config *config.Config
}

func NewFileService(biz *file.FileUsecase, config *config.Config) *FileService {
	return &FileService{biz: biz, config: config}
}

func (f *FileService) Upload(ctx context.Context, in *api.UploadFileRequest) (*api.UploadFileResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New(fmt.Errorf("not found metadata"), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "not_found_metadata")
	}
	identity := md.Get("x-identity")
	if len(identity) == 0 {
		return nil, errors.New(errors.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")
	}
	// in.Key = identity[0] + "/" + in.Key
	return f.biz.Upload(ctx, in, identity[0])
}

func (f *FileService) InitiateMultipartUpload(ctx context.Context, in *api.InitiateMultipartUploadRequest) (*api.InitiateMultipartUploadResponse, error) {
	return f.biz.InitiateUploadFile(ctx, in)
}
func (f *FileService) UploadMultipartFile(ctx context.Context, in *api.UploadMultipartFileRequest) (*api.UploadMultipartFileResponse, error) {
	return f.biz.UploadMultipartFileFile(ctx, in)
}
func (f *FileService) CompleteMultipartUpload(ctx context.Context, in *api.CompleteMultipartUploadRequest) (*api.CompleteMultipartUploadResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New(errors.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_metadata")
	}
	identity := md.Get("x-identity")
	if len(identity) == 0 {
		return nil, errors.New(errors.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")
	}
	return f.biz.CompleteMultipartUploadFile(ctx, in, identity[0])
}
func (f *FileService) AbortMultipartUpload(ctx context.Context, in *api.AbortMultipartUploadRequest) (*api.AbortMultipartUploadResponse, error) {
	return f.biz.AbortMultipartUpload(ctx, in)
}
func (f *FileService) Download(ctx context.Context, in *api.DownloadRequest) (*httpbody.HttpBody, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New(fmt.Errorf("not found metadata"), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "not_found_metadata")
	}
	identity := md.Get("x-identity")
	if len(identity) == 0 {
		return nil, errors.New(errors.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")
	}
	newKey, err := url.PathUnescape(in.Key)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_UNKNOWN), codes.InvalidArgument, "url_unescape")
	}
	in.Key = newKey
	buf, err := f.biz.Download(ctx, in, identity[0])
	if err != nil {
		return nil, err
	}

	shaer := sha256.New()
	shaer.Write(buf)
	rspMd := metadata.Pairs(
		gosdk.GetMetadataKey("Content-Length"), fmt.Sprintf("%d", len(buf)),
		gosdk.GetMetadataKey("x-content-sha256"), hex.EncodeToString(shaer.Sum(nil)),
	)
	err = grpc.SendHeader(ctx, rspMd)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_UNKNOWN), codes.Internal, "send_header")
	}

	rsp := &httpbody.HttpBody{
		ContentType: http.DetectContentType(buf),
		Data:        buf,
	}
	return rsp, err
}
func parseRangeHeader(rangeHeader string) (start, end int64, err error) {
	// 确保头部以"bytes="开头
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return 0, 0, fmt.Errorf("invalid range header: %s", rangeHeader)
	}

	// 去除"bytes="前缀
	rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeSpec, "-")
	if strings.HasPrefix(rangeSpec, "-") {
		start = 0
		end, err = strconv.ParseInt(strings.TrimPrefix(rangeSpec, "-"), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid start value: %s", parts[0])
		}
		return start, end, nil
	}
	if strings.HasSuffix(rangeSpec, "-") {
		end = 0
		start, err = strconv.ParseInt(strings.TrimSuffix(rangeSpec, "-"), 10, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid end value: %s", parts[1])

		}
		return start, end, nil
	}

	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range specification: %s", rangeSpec)
	}

	// 解析 start 值
	start, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start value: %s", parts[0])
	}

	// 解析 end 值
	end, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end value: %s", parts[1])
	}

	return start, end, nil
}
func (f *FileService) DownloadForRange(ctx context.Context, in *api.DownloadRequest) (*httpbody.HttpBody, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	var rangeStr string
	var start, end int64
	var err error
	if ok {
		if _, ok := md["range"]; !ok {
			return nil, errors.New(fmt.Errorf("range header not found"), int32(common.Code_PARAMS_ERROR), codes.InvalidArgument, "range_header_not_found")
		}
		rangeStr = md.Get("range")[0]
		start, end, err = parseRangeHeader(rangeStr)
		if err != nil {
			return nil, errors.New(err, int32(common.Code_UNKNOWN), codes.InvalidArgument, "parse_range_header")
		}
	}
	identity := GetIdentity(ctx)
	if identity == "" {
		return nil, errors.New(errors.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")

	}
	data, fileSize, err := f.biz.DownloadForRange(ctx, in, start, end, identity)
	if err != nil {
		return nil, err
	}
	if end <= 0 {
		end = fileSize
	}

	rspMd := metadata.Pairs(
		gosdk.GetMetadataKey("Content-Length"), fmt.Sprintf("%d", len(data)),
		gosdk.GetMetadataKey("Content-Range"), fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize),
		gosdk.GetMetadataKey("Accept-Ranges"), "bytes",
		"X-Http-Code", fmt.Sprintf("%d", http.StatusPartialContent),
	)
	err = grpc.SendHeader(ctx, rspMd)
	if err != nil {
		return nil, errors.New(err, int32(common.Code_UNKNOWN), codes.Internal, "send_header")
	}
	return &httpbody.HttpBody{
		ContentType: "application/octet-stream",
		Data:        data,
	}, nil
}
func (f *FileService) Delete(ctx context.Context, in *api.DeleteRequest) (*api.DeleteResponse, error) {
	identity := GetIdentity(ctx)
	if identity == "" {
		return nil, errors.New(errors.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")
	}
	return f.biz.Delete(ctx, in, identity)
}
func (f *FileService) Metadata(ctx context.Context, in *api.FileMetadataRequest) (*api.FileMetadataResponse, error) {
	identity := GetIdentity(ctx)
	if identity == "" {
		return nil, errors.New(errors.ErrIdentityMissing, int32(user.UserSvrCode_USER_IDENTITY_MISSING_ERR), codes.InvalidArgument, "not_found_identity")
	}
	rsp, err := f.biz.Metadata(ctx, in, identity)
	if err != nil {
		return nil, err

	}
	timestamp := time.UnixMilli(rsp.ModifyTime)
	lastModified := timestamp.UTC().Format(time.RFC1123)

	rspMd := metadata.Pairs(
		gosdk.GetMetadataKey("X-File-Name"), rsp.Name,
		gosdk.GetMetadataKey("content-type"), rsp.ContentType,
		gosdk.GetMetadataKey("Etag"), rsp.Etag,
		gosdk.GetMetadataKey("Last-Modified"), lastModified,
		gosdk.GetMetadataKey("Accept-Ranges"), "bytes",
		gosdk.GetMetadataKey("Content-Length"), fmt.Sprintf("%d", rsp.Size),
		gosdk.GetMetadataKey("X-File-Sha256"), rsp.Sha256,
		gosdk.GetMetadataKey("X-File-Version"), rsp.Version,
		gosdk.GetMetadataKey("Access-Control-Expose-Headers"), "Content-Length, Content-Range, Accept-Ranges, Last-Modified, ETag, Content-Type, X-File-name, X-File-Sha256",
	)
	// rspMd.Append(sdk.GetMetadataKey(), "Content-Length", "Content-Range", "Accept-Ranges", "Last-Modified", "ETag", "Content-Type", "x-file-name", "x-file-sha256")
	err = grpc.SendHeader(ctx, rspMd)
	if err != nil {

		return nil, errors.New(fmt.Errorf("非法的响应头,%w", err), int32(common.Code_UNKNOWN), codes.Internal, "send_header")

	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if httpMethod, ok := md["x-http-method"]; ok {
			if strings.EqualFold(httpMethod[0], "HEAD") {
				return nil, nil
			}
		}
	}

	return rsp, err
}
func (f *FileService) Desc() *grpc.ServiceDesc {
	return &api.FileService_ServiceDesc
}
