package gateway

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/begonia-org/begonia/internal/pkg/config"
	common "github.com/begonia-org/go-sdk/common/api/v1"
)

func readDesc(conf *config.Config) (ProtobufDescription, error) {
	desc := conf.GetLocalAPIDesc()
	log.Printf("read desc file:%s", desc)
	bin, err := os.ReadFile(desc)
	if err != nil {
		return nil, fmt.Errorf("read desc file error:%w", err)
	}
	pd, err := NewDescriptionFromBinary(bin, filepath.Dir(desc))
	if err != nil {
		return nil, err
	}
	err = pd.SetHttpResponse(common.E_HttpResponse)
	if err != nil {
		return nil, err
	}
	return pd, nil
}
func TestRegisterDynamicServices(t *testing.T) {
	// pd, _ := readDesc(config.NewConfig(cfg.ReadConfig("test")))
	// _ = &GatewayServer{}
	// gw.buildServiceDesc(pd)
}
