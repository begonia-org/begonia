package file

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
	api "github.com/begonia-org/go-sdk/api/v1"
	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc/metadata"
)

type TmpFile struct {
	sha256      string
	contentType string
	name        string
	path        string
	content     []byte
}
type FileUploadTestCase struct {
	tmp        *TmpFile
	useVersion bool
	key        string
	expectPath string
	title      string
	expectUri  string
	expectErr  error
}
type FileDownloadTestCase struct {
	uploadCase *FileUploadTestCase
	title      string
	exceptErr  error
}

func GenerateRandomFile(size int64) (*TmpFile, error) {
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

func initTestCase() *FileUsecase {
	config2 := cfg.ReadConfig("dev")
	configConfig := config.NewConfig(config2)
	// mySQLDao := data.NewMySQL(config2)
	// redisDao := data.NewRDB(config2)
	// etcdDao := data.NewEtcd(config2)
	// dataData := data.NewData(mySQLDao, redisDao, etcdDao)
	// fileRepo := data.NewFileRepoImpl(dataData)
	fileUsecase := NewFileUsecase(nil, configConfig)
	return fileUsecase
}
func TestUpload(t *testing.T) {
	useCase := initTestCase()
	tmp, err := GenerateRandomFile(1024 * 1024 * 4)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.path)

	// c.So(tmp, c.ShouldNotBeNil)
	ctx := context.Background()
	md := make(map[string]string)

	ctx = metadata.NewIncomingContext(ctx, metadata.New(md))
	cases := []FileUploadTestCase{
		{
			title:      "TestUpload without version",
			tmp:        tmp,
			useVersion: false,
			key:        "test/upload.test1",
			expectPath: filepath.Join(useCase.config.GetUploadDir(), "tester", "test", "upload.test1"),
			expectUri:  "test/upload.test1",
			expectErr:  nil,
		},
		{
			title:      "TestUpload with version",
			tmp:        tmp,
			useVersion: true,
			key:        "test/upload.test1",
			expectPath: filepath.Join(useCase.config.GetUploadDir(), "tester", "test", "upload.test1"),
			expectUri:  "test/upload.test1",
			expectErr:  nil,
		},
		{
			title:      "TestUpload with invalid key",
			tmp:        tmp,
			key:        "/test/upload.test3",
			expectPath: "",
			expectUri:  "",
			expectErr:  errors.ErrInvalidFileKey,
		},
	}
	for _, v := range cases {
		v := v
		c.Convey(v.title, t, func() {
			rsp, err := useCase.Upload(ctx, &api.UploadFileRequest{
				Key:         v.key,
				Content:     tmp.content,
				ContentType: tmp.contentType,
				Sha256:      tmp.sha256,
				UseVersion:  v.useVersion,
			}, "tester")
			if v.expectErr == nil {
				file, err := os.ReadFile(v.expectPath)
				c.So(err, c.ShouldBeNil)
				defer os.Remove(v.expectPath)
				c.So(file, c.ShouldResemble, tmp.content)
				c.So(rsp.Uri, c.ShouldEqual, v.expectUri)
				if v.useVersion {
					t.Log("version:", rsp.Version)
					c.So(rsp.Version, c.ShouldNotBeEmpty)
				}
			} else {
				c.So(err, c.ShouldNotBeNil)
				c.So(err.Error(), c.ShouldContainSubstring, v.expectErr.Error())
			}

		})
	}
}
func TestDownload(t *testing.T) {
	useCase := initTestCase()
	tmp, err := GenerateRandomFile(1024 * 1024 * 4)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.path)

	// c.So(tmp, c.ShouldNotBeNil)
	ctx := context.Background()
	md := make(map[string]string)

	ctx = metadata.NewIncomingContext(ctx, metadata.New(md))

	cases := []FileDownloadTestCase{
		{
			uploadCase: &FileUploadTestCase{
				title:      "TestDownload without version",
				tmp:        tmp,
				useVersion: false,
				key:        "test/upload.test",
				expectUri:  "test/upload.test",
			},
			title:     "TestDownload without version",
			exceptErr: nil,
		},
		{
			uploadCase: &FileUploadTestCase{
				title:      "TestDownload with version",
				tmp:        tmp,
				useVersion: true,
				key:        "test/upload.test",
				expectErr:  nil,
				expectUri:  "test/upload.test",
			},
			title:     "TestDownload with version",
			exceptErr: nil,
		},
		{
			uploadCase: &FileUploadTestCase{
				title:     "TestDownload with not found the key",
				tmp:       tmp,
				key:       "test/upload.test",
				expectUri: "test/no_upload.txt",
			},
			title:     "TestDownload with not found the key",
			exceptErr: fmt.Errorf("no such file or directory"),
		},
	}
	defer os.Remove(filepath.Join(useCase.config.GetUploadDir(), "versions", "tester", "test", "upload.test"))
	defer os.Remove(filepath.Join(useCase.config.GetUploadDir(), "tester", "test", "upload.test"))

	for _, v := range cases {
		v := v
		c.Convey(v.title, t, func() {
			rsp, err := useCase.Upload(ctx, &api.UploadFileRequest{
				Key:         v.uploadCase.key,
				Content:     tmp.content,
				ContentType: tmp.contentType,
				Sha256:      tmp.sha256,
				UseVersion:  v.uploadCase.useVersion,
			}, "tester")
			c.So(err, c.ShouldBeNil)
			time.Sleep(1 * time.Second)

			downloadRsp, err := useCase.Download(ctx, &api.DownloadRequest{
				Key:     v.uploadCase.expectUri,
				Version: rsp.Version,
			}, "tester")
			if v.exceptErr == nil {
				c.So(err, c.ShouldBeNil)
				c.So(downloadRsp, c.ShouldResemble, tmp.content)

				if v.uploadCase.useVersion {
					t.Log("version:", rsp.Version)
					c.So(rsp.Version, c.ShouldNotBeEmpty)
				}

			} else {
				c.So(err, c.ShouldNotBeNil)
				c.So(err.Error(), c.ShouldContainSubstring, v.exceptErr.Error())
			}

		})
	}
}

func TestMultipartFile(t *testing.T) {
	useCase := initTestCase()
	tmp, err := GenerateRandomFile(1024 * 1024 * 4)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.path)

	ctx := context.Background()
	md := make(map[string]string)

	ctx = metadata.NewIncomingContext(ctx, metadata.New(md))
	c.Convey("TestMultipartFile", t, func() {
		rsp, err := useCase.InitiateUploadFile(ctx, &api.InitiateMultipartUploadRequest{
			Key: "test/upload2.test",
		})
		uploadSize := 0
		c.So(err, c.ShouldBeNil)
		for i := 1; i <= 4; i++ {
			data := tmp.content[1024*1024*(i-1) : 1024*1024*i]
			hash := sha256.Sum256(data)
			uploadSize += len(data)
			// 将哈希值格式化为十六进制字符串
			hashString := fmt.Sprintf("%x", hash)

			_, err := useCase.UploadMultipartFileFile(ctx, &api.UploadMultipartFileRequest{
				Key:        "test/upload2.test",
				Content:    data,
				PartNumber: int64(i),
				UploadId:   rsp.UploadId,
				Sha256:     hashString,
			})
			c.So(err, c.ShouldBeNil)
		}
		cmpRsp, err := useCase.CompleteMultipartUploadFile(ctx, &api.CompleteMultipartUploadRequest{
			UploadId:    rsp.UploadId,
			Key:         "test/upload2.test",
			Sha256:      tmp.sha256,
			ContentType: tmp.contentType,
		}, "tester")
		c.So(err, c.ShouldBeNil)
		// t.Log(cmpRsp.Uri)
		_, err = useCase.Metadata(ctx, &api.FileMetadataRequest{
			Key:     cmpRsp.Uri,
			Version: cmpRsp.Sha256,
		}, "tester")
		c.So(err, c.ShouldBeNil)
		newFile := make([]byte, 0)
		downloadSize := 0
		for i := 1; i <= 4; i++ {
			data, _, err := useCase.DownloadForRange(ctx, &api.DownloadRequest{
				Key: "test/upload2.test",
			}, int64(i-1)*1024*1024, int64(i)*1024*1024, "tester")
			c.So(err, c.ShouldBeNil)
			downloadSize += len(data)
			newFile = append(newFile, data...)
		}
		c.So(uploadSize, c.ShouldEqual, len(tmp.content))
		c.So(downloadSize, c.ShouldEqual, uploadSize)

		hasher := sha256.New()
		hasher.Write(newFile)
		c.So(fmt.Sprintf("%x", hasher.Sum(nil)), c.ShouldEqual, tmp.sha256)
		// c.So(newFile, c.ShouldResemble, tmp.content)

	})

}

func TestMultipartFileWithVersion(t *testing.T) {
	useCase := initTestCase()
	tmp, err := GenerateRandomFile(1024 * 1024 * 4)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.path)

	ctx := context.Background()
	md := make(map[string]string)

	ctx = metadata.NewIncomingContext(ctx, metadata.New(md))
	c.Convey("TestMultipartFile with Version", t, func() {
		rsp, err := useCase.InitiateUploadFile(ctx, &api.InitiateMultipartUploadRequest{
			Key: "test/upload2.test",
		})
		uploadSize := 0
		c.So(err, c.ShouldBeNil)
		for i := 1; i <= 4; i++ {
			data := tmp.content[1024*1024*(i-1) : 1024*1024*i]
			hash := sha256.Sum256(data)
			uploadSize += len(data)
			// 将哈希值格式化为十六进制字符串
			hashString := fmt.Sprintf("%x", hash)

			_, err := useCase.UploadMultipartFileFile(ctx, &api.UploadMultipartFileRequest{
				Key:        "test/upload2.test",
				Content:    data,
				PartNumber: int64(i),
				UploadId:   rsp.UploadId,
				Sha256:     hashString,
			})
			c.So(err, c.ShouldBeNil)
		}
		cmpRsp, err := useCase.CompleteMultipartUploadFile(ctx, &api.CompleteMultipartUploadRequest{
			UploadId:    rsp.UploadId,
			Key:         "test/upload2.test",
			Sha256:      tmp.sha256,
			ContentType: tmp.contentType,
			UseVersion:  true,
		}, "tester")
		c.So(err, c.ShouldBeNil)
		c.So(cmpRsp.Version, c.ShouldNotBeEmpty)
		// t.Log(cmpRsp.Uri)
		_, err = useCase.Metadata(ctx, &api.FileMetadataRequest{
			Key:     cmpRsp.Uri,
			Version: cmpRsp.Sha256,
		}, "tester")
		c.So(err, c.ShouldBeNil)
		newFile := make([]byte, 0)
		downloadSize := 0
		for i := 1; i <= 4; i++ {
			data, _, err := useCase.DownloadForRange(ctx, &api.DownloadRequest{
				Key: "test/upload2.test",
			}, int64(i-1)*1024*1024, int64(i)*1024*1024, "tester")
			c.So(err, c.ShouldBeNil)
			downloadSize += len(data)
			newFile = append(newFile, data...)
		}
		c.So(uploadSize, c.ShouldEqual, len(tmp.content))
		c.So(downloadSize, c.ShouldEqual, uploadSize)

		hasher := sha256.New()
		hasher.Write(newFile)
		c.So(fmt.Sprintf("%x", hasher.Sum(nil)), c.ShouldEqual, tmp.sha256)
		// c.So(newFile, c.ShouldResemble, tmp.content)

	})

}
