package signature

import (
	"fmt"
	"net/url"
	"sort"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestSignature(t *testing.T) {
	c.Convey("Test Signature", t, func() {
		signer := NewAppAuthSigner("NWkbCslfh9ea2LjVIUsKehJuopPb65fn", "oVPNllSR1DfizdmdSF7wLjgABYbexdt4FZ1HWrI81dD5BeNhsyXpXPDFoDEyiSVe")
		uri, _ := url.Parse("http://127.0.0.1:12138/v1/gateway/info")
		req := &GatewayRequest{
			Headers: NewRequestHeader(),
			Method:  "GET",
			URL:     uri,
			Host:    "127.0.0.1:12138",
			Payload: nil,
		}
		err := signer.SignRequest(req)
		c.So(err, c.ShouldBeNil)
		t.Log(req.Headers.Get(HeaderXAuthorization))
	})
}
func TestSignature2(t *testing.T) {
	c.Convey("Test Signature", t, func() {
		// signer := NewAppAuthSigner("NWkbCslfh9ea2LjVIUsKehJuopPb65fn", "oVPNllSR1DfizdmdSF7wLjgABYbexdt4FZ1HWrI81dD5BeNhsyXpXPDFoDEyiSVe")
		// t.Log(signer.(*AppAuthSignerImpl).HexEncodeSHA256Hash([]byte("{}")))
		mimeTypes := []string{"application/json", "text/plain", "*/*"}
		sort.Strings(mimeTypes)
		fmt.Println(mimeTypes)
	})
}
