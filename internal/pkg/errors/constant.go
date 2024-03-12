package errors

import "errors"

var (
	ErrUserNotFound        = errors.New("用户不存在")
	ErrTokenInvalid        = errors.New("token无效")
	ErrTokenExpired        = errors.New("token过期")
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
	ErrRequestExpired      = errors.New("请求已过期")
)
