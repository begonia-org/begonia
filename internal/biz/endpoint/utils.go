package endpoint

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	gosdk "github.com/begonia-org/go-sdk"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/grpc/codes"
)

func deleteAll(ctx context.Context, pd gateway.ProtobufDescription) error {
	routersList := routers.Get()
	routersList.DeleteRouters(pd)
	gw := gateway.Get()
	gw.DeleteLoadBalance(pd)
	err := gateway.Get().DeleteHandlerClient(ctx, pd)
	return err
}

func getDescriptorSet(config *config.Config, key string, value []byte) (gateway.ProtobufDescription, error) {
	key = getEndpointId(config, key)
	outDir := config.GetGatewayDescriptionOut()
	pd, err := gateway.NewDescriptionFromBinary(value, filepath.Join(outDir, key))
	if err != nil {
		return nil, gosdk.NewError(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "new_description_from_binary")
	}
	return pd, nil
}
func getEndpointId(config *config.Config, key string) string {
	prefix := config.GetEndpointsPrefix()
	key = strings.TrimPrefix(key, prefix)
	key = strings.TrimPrefix(key, "/")
	key = strings.TrimSuffix(key, "/descriptor_set")
	return key
}
