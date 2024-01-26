package runtime

import "testing"

func TestLoader(t *testing.T) {

	load := NewProtoLoaderImpl(nil)
	err := load.LoadProto("protos.zip", "tmp", "./api/v1", "github.com/wetrycode/example")
	if err != nil {
		t.Error(err)
	}
}
