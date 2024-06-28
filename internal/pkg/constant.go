package pkg

import (
	"errors"
	"fmt"
)

var (
	ErrUserNotFound        = errors.New("用户不存在")
	ErrUserDisabled        = errors.New("用户已禁用")
	ErrTokenInvalid        = errors.New("token无效")
	ErrTokenExpired        = fmt.Errorf("token过期")
	ErrTokenMissing        = errors.New("authorization缺失")
	ErrUserPasswordInvalid = errors.New("密码错误")
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

	ErrUploadNotInitiate  = errors.New("上传未初始化")
	ErrSHA256NotMatch     = errors.New("sha256不匹配")
	ErrUploadIdMissing    = errors.New("uploadId缺失")
	ErrUploadIdNotFound   = errors.New("uploadId未找到")
	ErrPartNumberMissing  = errors.New("partNumber缺失")
	ErrInvalidFileKey     = errors.New("无效的文件路径")
	ErrFileKeyMissing     = errors.New("file key 缺失")
	ErrInvalidRange       = errors.New("无效的range")
	ErrBucketNotFound     = errors.New("bucket不存在")
	ErrFsEngineNotSupport = errors.New("文件引擎不支持")

	ErrIdentityMissing = errors.New("identity缺失")

	ErrUnknownLoadBalancer = errors.New("未知的负载均衡器")

	ErrAPIKeyNotMatch = errors.New("api key不匹配")

	ErrEndpointExists = errors.New("endpoint已存在")

	ErrEndpointNotExists = errors.New("endpoint不存在")
)
