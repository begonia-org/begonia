package integration_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/go-sdk/client"
	common "github.com/begonia-org/go-sdk/common/api/v1"
	c "github.com/smartystreets/goconvey/convey"
)

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

func upload(t *testing.T) {
	c.Convey("test upload file", t, func() {
		// test upload file
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret)
		_, filename, _, _ := runtime.Caller(0)

		pbFile := filepath.Join(filepath.Dir(filename), "testdata", "helloworld.pb")
		srcSha256, _ := sumFileSha256(pbFile)
		rsp, err := apiClient.UploadFile(context.Background(), pbFile, "test/helloworld.pb", true)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.Uri, c.ShouldNotBeEmpty)
		conf := cfg.NewConfig(config.ReadConfig(begonia.Env))

		saveDir := filepath.Join(conf.GetUploadDir(), filepath.Dir(rsp.Uri))
		filename = filepath.Base(rsp.Uri)
		filePath := filepath.Join(saveDir, filename)

		_, err = os.Stat(filePath)
		c.So(err, c.ShouldBeNil)
		dstSha256, err := sumFileSha256(filePath)
		c.So(err, c.ShouldBeNil)
		c.So(srcSha256, c.ShouldEqual, dstSha256)
		c.So(rsp.Version, c.ShouldNotBeEmpty)
		// c.So(info.Size(), c., 102)
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
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret)
		var err error
		tmpFile, err = generateRandomFile(1024 * 1024 * 20)
		c.So(err, c.ShouldBeNil)
		defer os.Remove(tmpFile.path)
		rsp, err := apiClient.UploadFileWithMuiltParts(context.Background(), tmpFile.path, "test/tmp.bin", true)
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		c.So(rsp.Uri, c.ShouldNotBeEmpty)
		c.So(rsp.Version, c.ShouldNotBeEmpty)
		// c.So(rsp.Sha256,c.ShouldEqual,tmp.sha256)
		env:=begonia.Env
		if env==""{
			env="test"
		}
		conf := cfg.NewConfig(config.ReadConfig(env))

		filePath := filepath.Join(conf.GetUploadDir(), rsp.Uri)
		// filename := filepath.Base(rsp.Uri)
		// filePath := filepath.Join(saveDir, filename)

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
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret)
		tmp, err := os.CreateTemp("", "testfile-*.txt")
		c.So(err, c.ShouldBeNil)
		defer tmp.Close()
		defer os.Remove(tmp.Name())
		sha256Str, err := apiClient.DownloadFile(context.Background(), sdkAPPID + "/test/helloworld.pb", tmp.Name(), "")
		c.So(err, c.ShouldBeNil)
		downloadedSha256, _ := sumFileSha256(tmp.Name())
		t.Log(sha256Str)
		c.So(sha256Str, c.ShouldEqual, downloadedSha256)
	})
}
func downloadParts(t *testing.T) {
	c.Convey("test download parts file", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret)
		tmp, err := os.CreateTemp("", "testfile-*.txt")
		c.So(err, c.ShouldBeNil)
		defer tmp.Close()
		defer os.Remove(tmp.Name())
		rsp, err := apiClient.DownloadMultiParts(context.Background(), sdkAPPID+"/test/tmp.bin", tmp.Name(), "")
		c.So(err, c.ShouldBeNil)
		// c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		downloadedSha256, _ := sumFileSha256(tmp.Name())
		
		c.So(rsp.Sha256, c.ShouldEqual, downloadedSha256)

	})
}
func deleteFile(t *testing.T) {
	c.Convey("test delete file", t, func() {
		apiClient := client.NewFilesAPI(apiAddr, accessKey, secret)
		rsp, err := apiClient.DeleteFile(context.Background(), sdkAPPID +"/test/helloworld.pb")
		c.So(err, c.ShouldBeNil)
		c.So(rsp.StatusCode, c.ShouldEqual, common.Code_OK)
		conf := cfg.NewConfig(config.ReadConfig("test"))

		saveDir := filepath.Join(conf.GetUploadDir(), filepath.Dir("test/helloworld.pb"))
		filename := filepath.Base("test/helloworld.pb")
		filePath := filepath.Join(saveDir, filename)

		_, err = os.Stat(filePath)
		// c.So(err, c.ShouldNotBeNil)
		c.So(os.IsNotExist(err), c.ShouldBeTrue)

	})
}
func TestFile(t *testing.T) {
	t.Run("upload", upload)
	t.Run("download", download)
	t.Run("uploadParts", uploadParts)
	t.Run("downloadParts", downloadParts)
	t.Run("deleteFile", deleteFile)
}
