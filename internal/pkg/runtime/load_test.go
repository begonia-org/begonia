package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cfg "github.com/wetrycode/begonia/config"
	"github.com/wetrycode/begonia/internal/pkg/config"

	c "github.com/smartystreets/goconvey/convey" // 别名导入
)

func TestLoader(t *testing.T) {
	c.Convey("test loader", t, func() {
		load := NewProtoLoaderImpl(config.NewConfig(
			cfg.ReadConfig("dev"),
		))
		err := load.LoadProto("protos.zip", "github.com/wetrycode/example", "./api/v1", "example")
		if err != nil {
			t.Error(err)
		}
		conf := cfg.ReadConfig("dev")
		pluginDir := conf.GetString("endpoints.plugins.dir")
		pluginDir = filepath.Join(pluginDir, "example")
		pluginName := strings.ReplaceAll("github.com/wetrycode/example", "/", ".")
		pluginPath:=filepath.Join(pluginDir, fmt.Sprintf("%s.so", pluginName))
		fs, err := os.Stat(pluginPath)
		c.So(os.IsExist(err), c.ShouldBeFalse)
		t.Log(fs.Name())
		defer os.RemoveAll(pluginDir)
	})

}
