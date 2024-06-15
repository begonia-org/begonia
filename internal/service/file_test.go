package service_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/data"

	"github.com/begonia-org/begonia/internal/pkg"

	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/service"
	api "github.com/begonia-org/go-sdk/api/file/v1"
	"github.com/begonia-org/go-sdk/client"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc/metadata"
)

var fileBucket = ""

func sumFileSha256(src string) (string, error) {
	file, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
func newFileUsecases() map[string]file.FileUsecase {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	cnf := config.ReadConfig(env)
	conf := cfg.NewConfig(cnf)
	repo := data.NewFileRepo(cnf, gateway.Log)
	return file.NewFileUsecase(conf, repo)
}
func makeBucket(t *testing.T) {
	c.Convey("test make bucket", t, func() {
		fileBucket = fmt.Sprintf("test-service-bucket-%s", time.Now().Format("20060102150405"))
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		rsp, err := apiClient.CreateBucket(context.Background(), fileBucket, "test", false)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		minioFile := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_MINIO)
		rsp, err = minioFile.CreateBucket(context.Background(), fileBucket, "test", false)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)

	})
}
func upload(t *testing.T) {
	env := begonia.Env
	if env == "" {
		env = "dev"
	}
	c.Convey("test upload file", t, func() {
		// test upload file
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		_, filename, _, _ := runtime.Caller(0)

		pbFile := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(filename))), "testdata", "helloworld.pb")
		srcSha256, _ := sumFileSha256(pbFile)
		rsp, err := apiClient.UploadFile(context.Background(), pbFile, "test/helloworld.pb", fileBucket, true)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.Uri, c.ShouldNotBeEmpty)
		conf := cfg.NewConfig(config.ReadConfig(env))

		filePath := filepath.Join(conf.GetUploadDir(), fileBucket, rsp.Uri)

		_, err = os.Stat(filePath)
		c.So(err, c.ShouldBeNil)
		dstSha256, err := sumFileSha256(filePath)
		c.So(err, c.ShouldBeNil)
		c.So(srcSha256, c.ShouldEqual, dstSha256)
		c.So(rsp.Version, c.ShouldNotBeEmpty)
	})
}

type TmpFile struct {
	sha256      string
	contentType string
	name        string
	path        string
	content     []byte
}

var tmpFile *TmpFile

func generateRandomFile(size int64) (*TmpFile, error) {
	// 创建临时文件
	tmp, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		return nil, err
	}
	defer tmp.Close()

	// 创建一个用于计算SHA256的hash.Writer
	hasher := sha256.New()

	// 创建一个MultiWriter，同时向文件和hasher写入数据
	writer := io.MultiWriter(tmp, hasher)

	// 使用crypto/rand生成随机数据，并填充文件至指定大小
	if _, err = io.CopyN(writer, rand.Reader, size); err != nil {
		return nil, err
	}

	// 计算哈希值
	hashValue := fmt.Sprintf("%x", hasher.Sum(nil))

	// 重新打开文件以读取其内容用于检测Content-Type
	fileContent, err := os.ReadFile(tmp.Name())
	if err != nil {
		return nil, err
	}

	// 检测Content-Type
	contentType := http.DetectContentType(fileContent)
	return &TmpFile{
		sha256:      hashValue,
		contentType: contentType,
		name:        tmp.Name(),
		path:        tmp.Name(),
		content:     fileContent,
	}, nil
	// return tmpfile.Name(), hashValue, contentType, nil
}
func uploadParts(t *testing.T) {
	c.Convey("test upload file", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		var err error
		tmpFile, err = generateRandomFile(1024 * 1024 * 2)
		c.So(err, c.ShouldBeNil)
		defer os.Remove(tmpFile.path)
		rsp, err := apiClient.UploadFileWithMuiltParts(context.Background(), tmpFile.path, "test/tmp.bin", fileBucket, true)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.Uri, c.ShouldNotBeEmpty)
		c.So(rsp.Version, c.ShouldNotBeEmpty)
		// c.So(rsp.Sha256,c.ShouldEqual,tmp.sha256)
		env := "dev"
		if begonia.Env != "" {
			env = "test"
		}
		conf := cfg.NewConfig(config.ReadConfig(env))

		filePath := filepath.Join(conf.GetUploadDir(), fileBucket, rsp.Uri)

		file, err := os.Open(filePath)
		c.So(err, c.ShouldBeNil)
		defer file.Close()

		hasher := sha256.New()
		buf := make([]byte, 32*1024) // 32KB buffer

		for {
			n, err := file.Read(buf)
			// c.So(err, c.ShouldNotEqual, io.EOF)
			if err != nil && err != io.EOF {
				c.So(err.Error(), c.ShouldEqual, io.EOF.Error())
				break
			}
			if n == 0 {
				break
			}

			// Update the hash with the chunk read
			hasher.Write(buf[:n])
		}

		sum := hasher.Sum(nil)
		c.So(hex.EncodeToString(sum), c.ShouldEqual, tmpFile.sha256)
	})
}
func download(t *testing.T) {
	c.Convey("test download file", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		tmp, err := os.CreateTemp("", "testfile-*.txt")
		c.So(err, c.ShouldBeNil)
		defer tmp.Close()
		defer os.Remove(tmp.Name())
		sha256Str, err := apiClient.DownloadFile(context.Background(), sdkAPPID+"/test/helloworld.pb", tmp.Name(), "", fileBucket)
		c.So(err, c.ShouldBeNil)
		_, err = os.Stat(tmp.Name())
		c.So(err, c.ShouldBeNil)
		downloadedSha256, err := sumFileSha256(tmp.Name())
		c.So(err, c.ShouldBeNil)
		t.Log(sha256Str)
		c.So(sha256Str, c.ShouldEqual, downloadedSha256)
	})
}
func downloadParts(t *testing.T) {
	c.Convey("test download parts file", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		tmp, err := os.CreateTemp("", "testfile-*.txt")
		c.So(err, c.ShouldBeNil)
		defer tmp.Close()
		defer os.Remove(tmp.Name())
		rsp, err := apiClient.DownloadMultiParts(context.Background(), sdkAPPID+"/test/tmp.bin", tmp.Name(), "", fileBucket)
		c.So(err, c.ShouldBeNil)
		// c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		downloadedSha256, _ := sumFileSha256(tmp.Name())

		c.So(rsp.Sha256, c.ShouldEqual, downloadedSha256)

	})
}
func deleteFile(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	c.Convey("test delete file", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		rsp, err := apiClient.DeleteFile(context.Background(), sdkAPPID+"/test/helloworld.pb", fileBucket)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		conf := cfg.NewConfig(config.ReadConfig(env))

		saveDir := filepath.Join(conf.GetUploadDir(), filepath.Dir("test/helloworld.pb"))
		filename := filepath.Base("test/helloworld.pb")
		filePath := filepath.Join(saveDir, filename)

		_, err = os.Stat(filePath)
		// c.So(err, c.ShouldNotBeNil)
		c.So(os.IsNotExist(err), c.ShouldBeTrue)

	})
}
func testRangeDownload(t *testing.T) {
	c.Convey("test range download file", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		tmp, err := os.CreateTemp("", "testfile-*.txt")
		c.So(err, c.ShouldBeNil)
		defer tmp.Close()
		defer os.Remove(tmp.Name())
		rsp, err := apiClient.RangeDownload(context.Background(), sdkAPPID+"/test/tmp.bin", "", -1, 128, fileBucket)
		c.So(err, c.ShouldBeNil)
		c.So(len(rsp), c.ShouldEqual, 129)

		rsp, err = apiClient.RangeDownload(context.Background(), sdkAPPID+"/test/tmp.bin", "", 128, -1, fileBucket)
		c.So(err, c.ShouldBeNil)
		c.So(len(rsp), c.ShouldEqual, 1024*1024*2-128)

		patch := gomonkey.ApplyFuncReturn(service.GetIdentity, "")
		defer patch.Reset()
		_, err = apiClient.RangeDownload(context.Background(), sdkAPPID+"/test/tmp.bin", "", 128, -1, fileBucket)
		c.So(err, c.ShouldNotBeNil)
		patch.Reset()

	})
}

func testUploadErr(t *testing.T) {
	c.Convey("test upload file err", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		srv := service.NewFileSvrForTest(cnf, gateway.Log)
		_, err := srv.Upload(context.Background(), &api.UploadFileRequest{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrIdentityMissing.Error())

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("app_id", sdkAPPID))
		_, err = srv.Upload(ctx, &api.UploadFileRequest{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrIdentityMissing.Error())

	})
}
func testDownloadErr(t *testing.T) {
	c.Convey("test download file err", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		srv := service.NewFileSvrForTest(cnf, gateway.Log)
		_, err := srv.Download(context.Background(), &api.DownloadRequest{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrIdentityMissing.Error())

		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("app_id", sdkAPPID))
		_, err = srv.Download(ctx, &api.DownloadRequest{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrIdentityMissing.Error())

		patch := gomonkey.ApplyFuncReturn(url.PathUnescape, "", fmt.Errorf("test PathUnescape error"))
		defer patch.Reset()
		ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-identity", sdkAPPID))
		_, err = srv.Download(ctx, &api.DownloadRequest{Key: "test"})
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test PathUnescape error")

	})
}
func testRangeDownloadErr(t *testing.T) {
	c.Convey("test range download file err", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		srv := service.NewFileSvrForTest(cnf, gateway.Log)

		cases := []struct {
			rangeStr string
			err      error
		}{
			{
				rangeStr: "",
				err:      fmt.Errorf("range header not found"),
			},
			{
				rangeStr: "0-0",
				err:      fmt.Errorf("invalid range header"),
			},
			{
				rangeStr: "bytes=-ffee",
				err:      fmt.Errorf("invalid end value"),
			},
			{
				rangeStr: "bytes=ffee-",
				err:      fmt.Errorf("invalid start value"),
			},
			{
				rangeStr: "bytes=1024",
				err:      fmt.Errorf("invalid range specification"),
			},
			{
				rangeStr: "bytes=tgg-1024",
				err:      fmt.Errorf("invalid start value"),
			},
			{
				rangeStr: "bytes=1024-tgg",
				err:      fmt.Errorf("invalid end value"),
			},
		}
		for _, cs := range cases {
			md := metadata.New(nil)
			md.Set("x-identity", sdkAPPID)
			if cs.rangeStr != "" {
				md.Set("range", cs.rangeStr)
			}
			ctx := metadata.NewIncomingContext(context.Background(), md)
			_, err := srv.DownloadForRange(ctx, &api.DownloadRequest{Key: sdkAPPID + "/test/helloworld.pb"})
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, cs.err.Error())
		}
		fileUseCase := newFileUsecases()

		patch := gomonkey.ApplyMethodReturn(fileUseCase[api.FileEngine_FILE_ENGINE_LOCAL.String()], "DownloadForRange", nil, int64(0), fmt.Errorf("test download for range error"))
		defer patch.Reset()
		md := metadata.New(nil)
		md.Set("x-identity", sdkAPPID)
		md.Set("range", "bytes=0-0")
		ctx := metadata.NewIncomingContext(context.Background(), md)
		_, err := srv.DownloadForRange(ctx, &api.DownloadRequest{Key: sdkAPPID + "/test/helloworld.pb", Engine: api.FileEngine_FILE_ENGINE_LOCAL.String(), Bucket: fileBucket})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test download for range error")
		patch.Reset()

	})
}

func testAbortUpload(t *testing.T) {
	c.Convey("test range download file", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		_, err := apiClient.AbortUpload(context.Background(), "test/tmp.bindddasd")
		c.So(err, c.ShouldNotBeNil)

	})
}
func testDelErr(t *testing.T) {
	c.Convey("test delete file err", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		srv := service.NewFileSvrForTest(cnf, gateway.Log)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("app_id", sdkAPPID))
		_, err := srv.Delete(ctx, &api.DeleteRequest{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrIdentityMissing.Error())

	})
}
func testMetaErr(t *testing.T) {
	c.Convey("test meta file err", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		srv := service.NewFileSvrForTest(cnf, gateway.Log)
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("app_id", sdkAPPID))
		_, err := srv.Metadata(ctx, &api.FileMetadataRequest{})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrIdentityMissing.Error())
		fileUseCase := newFileUsecases()

		patch := gomonkey.ApplyMethodReturn(fileUseCase[api.FileEngine_FILE_ENGINE_LOCAL.String()], "Metadata", nil, fmt.Errorf("test metadata error"))
		defer patch.Reset()
		ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-identity", sdkAPPID))
		_, err = srv.Metadata(ctx, &api.FileMetadataRequest{Key: sdkAPPID + "/test/helloworld.pb", Bucket: fileBucket, Engine: api.FileEngine_FILE_ENGINE_LOCAL.String()})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "test metadata error")

	})
}
func testFileList(t *testing.T) {
	c.Convey("test file list", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		rsp, err := apiClient.ListFiles(context.Background(), api.FileEngine_FILE_ENGINE_LOCAL.String(), fileBucket, 1, 10)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
	})
	c.Convey("test file list err", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret, api.FileEngine_FILE_ENGINE_LOCAL)
		fileUseCase := newFileUsecases()
		patch := gomonkey.ApplyMethodReturn(fileUseCase[api.FileEngine_FILE_ENGINE_LOCAL.String()], "List", nil, fmt.Errorf("test metadata error"))
		defer patch.Reset()
		_, err := apiClient.ListFiles(context.Background(), api.FileEngine_FILE_ENGINE_LOCAL.String(), fileBucket, 1, 10)
		c.So(err, c.ShouldNotBeNil)
		patch.Reset()
		patch2 := gomonkey.ApplyFuncReturn(service.GetIdentity, "")
		defer patch2.Reset()
		_, err = apiClient.ListFiles(context.Background(), api.FileEngine_FILE_ENGINE_LOCAL.String(), fileBucket, 1, 10)
		c.So(err, c.ShouldNotBeNil)
		patch2.Reset()

		_, err = apiClient.ListFiles(context.Background(), "api.FileEngine_FILE_ENGINE_LOCAL.String()", fileBucket, 1, 10)
		c.So(err, c.ShouldNotBeNil)

	})
}
func TestFile(t *testing.T) {
	t.Run("makeBucket", makeBucket)
	t.Run("upload", upload)
	t.Run("download", download)
	t.Run("testUploadErr", testUploadErr)
	t.Run("testDownloadErr", testDownloadErr)
	t.Run("uploadParts", uploadParts)
	t.Run("testRangeDownload", testRangeDownload)
	t.Run("testFileList", testFileList)
	t.Run("testAbortUpload", testAbortUpload)
	t.Run("testRangeDownloadErr", testRangeDownloadErr)
	t.Run("downloadParts", downloadParts)
	t.Run("deleteFile", deleteFile)
	t.Run("testDelErr", testDelErr)
	t.Run("testMetaErr", testMetaErr)
}
