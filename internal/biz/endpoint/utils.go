package endpoint

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	"github.com/begonia-org/begonia/internal/pkg/routers"
	"github.com/begonia-org/begonia/transport"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	"google.golang.org/grpc/codes"
)

func deleteAll(ctx context.Context, pd transport.ProtobufDescription) error {
	routersList := routers.Get()
	routersList.DeleteRouters(pd)
	gw := transport.Get()
	gw.DeleteLoadBalance(pd)
	err := transport.Get().DeleteHandlerClient(ctx, pd)
	return err
}

func getDescriptorSet(config *config.Config, key string, value []byte) (transport.ProtobufDescription, error) {
	key = getEndpointId(config, key)
	outDir := config.GetGatewayDescriptionOut()
	pd, err := transport.NewDescriptionFromBinary(value, filepath.Join(outDir, key))
	if err != nil {
		return nil, errors.New(err, int32(common.Code_INTERNAL_ERROR), codes.Internal, "new_description_from_binary")
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
