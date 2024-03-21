package errors

import (
	"errors"
	"fmt"
	"runtime"

	common "github.com/begonia-org/begonia/common/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

type SrvError struct {
	Err      error
	ErrCode  int32
	GrpcCode codes.Code
	Action   string
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}
func As(err error, target interface{}) bool {
	return errors.As(err, target)

}
func New(err error, code int32, grpcCode codes.Code, action string) error {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}
	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
	}


	srvErr := &common.Errors{
		Code:    code,
		Message: err.Error(),
		Action:  action,
		File:    file,
		Line:    int32(line),
		Fn:      funcName,
	}
	st := status.New(grpcCode, err.Error())
	detailProto, err := anypb.New(srvErr)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to marshal error details: %v", err)
	}
	st, err = st.WithDetails(detailProto)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to marshal error details: %v", err)

	}
	return st.Err()
	// srvErr:=status.New(0,err.Error())
	// srvErr.WithDetails()
	// return nil
}
func (s *SrvError) Error() string {
	return fmt.Sprintf("%s|%d", s.Err.Error(), s.ErrCode)
}
func (s *SrvError) Code() int32 {
	return s.ErrCode
}

var (
	ErrUserNotFound        = errors.New("用户不存在")
	ErrTokenInvalid        = errors.New("token无效")
	ErrTokenExpired        = fmt.Errorf("token过期")
	ErrTokenMissing        = errors.New("authorization缺失")
	ErrTokenBlackList      = errors.New("token在黑名单中")
	ErrNoMetadata          = errors.New("获取metadata失败")
	ErrAuthDecrypt         = errors.New("解密失败")
	ErrEncrypt             = errors.New("加密失败")
	ErrDecode              = errors.New("解码失败")
	ErrEncode              = errors.New("编码失败")
	ErrRemoveBlackList     = errors.New("移除黑名单失败")
	ErrCacheBlacklist      = errors.New("缓存黑名单失败")
	ErrUidMissing          = errors.New("uid缺失")
	ErrTokenNotActive      = errors.New("token未生效")
	ErrTokenIssuer         = errors.New("token签发者错误")
	ErrHeaderTokenFormat   = errors.New("header中token格式错误")
	ErrAppSignatureMissing = errors.New("app签名缺失")
	ErrAppSignatureInvalid = errors.New("app签名无效")
	ErrAppAccessKeyMissing = errors.New("app access key缺失")
	ErrAppXDateMissing     = errors.New("app x-date缺失")
	ErrRequestExpired      = errors.New("请求已过期")

	ErrUploadNotInitiate = errors.New("上传未初始化")
	ErrSHA256NotMatch    = errors.New("sha256不匹配")
	ErrUploadIdMissing   = errors.New("uploadId缺失")
	ErrUploadIdNotFound  = errors.New("uploadId未找到")
	ErrPartNumberMissing = errors.New("partNumber缺失")
	ErrInvalidFileKey    = errors.New("无效的文件路径")
	ErrFileKeyMissing    = errors.New("file key 缺失")

)
