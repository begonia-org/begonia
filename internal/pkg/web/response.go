package web

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/spark-lence/tiga"
	srvErr "github.com/spark-lence/tiga/errors"
	_ "github.com/wetrycode/begonia/api/v1"
	"github.com/wetrycode/begonia/internal/pkg/config"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	common "github.com/wetrycode/begonia/common/api/v1"
)

func unwrap(err error) *srvErr.Errors {
	var se *srvErr.Errors
	if errors.As(err, &se) {
		return se
	}
	return nil

}
func MakeResponse(data protoreflect.ProtoMessage, srcErr error) (*common.APIResponse, error) {
	message := "Internal Error"
	code := int32(common.Code_INTERNAL_ERROR.Number())
	if srcErr != nil {
		se := unwrap(srcErr)
		if se != nil {
			message = se.ClientMessage()
			code = se.Code()
		}
	} else {
		code = int32(common.Code_OK.Number())
		message = "ok"
	}
	rsp, err := tiga.MakeResponse(code, data, srcErr, message, fmt.Sprintf("%s.APIResponse", config.APIPkg))
	if rsp != nil {
		// 序列化 *dynamicapi.Message
		serializedMsg, mErr := proto.Marshal(rsp)
		if mErr != nil {
			return nil, fmt.Errorf("序列化响应失败,%w", mErr) // 处理错误
			// 处理错误
		}
		// 反序列化为 common.APIResponse
		var apiResponse *common.APIResponse = &common.APIResponse{}
		mErr = proto.Unmarshal(serializedMsg, apiResponse)
		if mErr != nil {
			return nil, fmt.Errorf("反序列化响应失败,%w", mErr) // 处理错误
			// 处理错误
		}
		return apiResponse, err
	}
	return nil, err
}
