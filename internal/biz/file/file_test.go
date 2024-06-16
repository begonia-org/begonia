package file_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal/biz/file"
	"github.com/begonia-org/begonia/internal/data"
	"github.com/begonia-org/begonia/internal/pkg"
	api "github.com/begonia-org/go-sdk/api/file/v1"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spark-lence/tiga"

	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	c "github.com/smartystreets/goconvey/convey"
)

type tmpFile struct {
	sha256      string
	contentType string
	name        string
	path        string
	content     []byte
}
type fileUploadTestCase struct {
	tmp        *tmpFile
	useVersion bool
	key        string
	expectPath string
	title      string
	author     string
	bucket     string
	expectUri  string
	expectErr  error
}
type FileDownloadTestCase struct {
	// uploadCase *fileUploadTestCase
	title      string
	key        string
	exceptErr  error
	useVersion bool
	version    string
	author     string
	sha256     string
}

var fileAuthor = ""
var fileSha256 = ""
var uploadId = ""
var bucket = ""
var versions map[string]string = map[string]string{}
var tmp *tmpFile = nil
var tmp3 *tmpFile = nil
var bigFileSha256 = ""

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
func generateRandomFile(size int64) (*tmpFile, error) {
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
	return &tmpFile{
		sha256:      hashValue,
		contentType: contentType,
		name:        tmp.Name(),
		path:        tmp.Name(),
		content:     fileContent,
	}, nil
	// return tmpfile.Name(), hashValue, contentType, nil
}

func newFileBiz() *file.FileUsecaseImpl {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	cnf := config.ReadConfig(env)
	conf := cfg.NewConfig(cnf)
	repo := data.NewFileRepo(cnf, gateway.Log)
	return file.NewLocalFileUsecase(conf, repo)
}
func testMkBucket(t *testing.T) {
	fileBiz := newFileBiz()
	snk, _ := tiga.NewSnowflake(1)

	fileAuthor = snk.GenerateIDString()
	bucket = fmt.Sprintf("bucket-biz-%s", time.Now().Format("20060102150405"))
	c.Convey("test make bucket success", t, func() {
		rsp, err := fileBiz.MakeBucket(context.TODO(), &api.MakeBucketRequest{
			Bucket: bucket,
		})
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
	})
	c.Convey("test make bucket fail", t, func() {
		patch := gomonkey.ApplyFuncReturn(os.MkdirAll, fmt.Errorf("mkdir error"))
		defer patch.Reset()
		_, err := fileBiz.MakeBucket(context.TODO(), &api.MakeBucketRequest{
			Bucket: "",
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "mkdir error")
	})
}
func testPutFile(t *testing.T) {
	fileBiz := newFileBiz()
	var err error
	tmp, err = generateRandomFile(1024 * 1024 * 1)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp.path)

	tmp2, err := generateRandomFile(1024 * 1024 * 1)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp2.path)

	tmp3, err = generateRandomFile(1024 * 1024 * 1)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tmp3.path)
	tmp2.sha256 = "deffweferfwcerfreverver"
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	fileSha256 = tmp.sha256
	// bucket = fmt.Sprintf("bucket-biz-%s", time.Now().Format("20060102150405"))

	cases := []fileUploadTestCase{
		{
			title:      "test upload without version",
			tmp:        tmp,
			useVersion: false,
			key:        "test/upload.test1",
			expectPath: filepath.Join(cnf.GetUploadDir(), bucket, fileAuthor, "test", "upload.test1"),
			expectUri:  fileAuthor + "/test/upload.test1",
			expectErr:  nil,
			author:     fileAuthor,
			bucket:     bucket,
		},
		{
			title:      "test upload with version",
			tmp:        tmp,
			useVersion: true,
			key:        "test/upload.test2",
			expectPath: filepath.Join(cnf.GetUploadDir(), bucket, fileAuthor, "test", "upload.test2"),
			expectUri:  fileAuthor + "/test/upload.test2",
			expectErr:  nil,
			author:     fileAuthor,
			bucket:     bucket,
		},
		{
			title:      "test upload with version1",
			tmp:        tmp,
			useVersion: true,
			key:        "test/upload.test2",
			expectPath: filepath.Join(cnf.GetUploadDir(), bucket, fileAuthor, "test", "upload.test2"),
			expectUri:  fileAuthor + "/test/upload.test2",
			expectErr:  nil,
			author:     fileAuthor,
			bucket:     bucket,
		},
		{
			title:      "test upload with version2",
			tmp:        tmp3,
			useVersion: true,
			key:        "test/upload.test2",
			expectPath: filepath.Join(cnf.GetUploadDir(), bucket, fileAuthor, "test", "upload.test2"),
			expectUri:  fileAuthor + "/test/upload.test2",
			expectErr:  nil,
			author:     fileAuthor,
			bucket:     bucket,
		},
		{
			title:      "test upload with version3",
			tmp:        tmp,
			useVersion: true,
			key:        "test/upload.test2",
			expectPath: filepath.Join(cnf.GetUploadDir(), bucket, fileAuthor, "test", "upload.test2"),
			expectUri:  fileAuthor + "/test/upload.test2",
			expectErr:  nil,
			author:     fileAuthor,
			bucket:     bucket,
		},

		{
			title:      "test upload with invalid key",
			tmp:        tmp,
			key:        "/test/upload.test3",
			expectPath: "",
			expectUri:  "",
			expectErr:  pkg.ErrInvalidFileKey,
			author:     fileAuthor,
			bucket:     bucket,
		},
		{
			title:      "test upload with invalid author",
			tmp:        tmp,
			key:        "test/upload.test4",
			expectPath: "",
			expectUri:  "",
			expectErr:  pkg.ErrIdentityMissing,
			author:     "",
			bucket:     bucket,
		},
		{
			title:      "test upload fail with not match sha256",
			tmp:        tmp2,
			key:        "test/upload.test7",
			expectPath: "",
			expectUri:  "",
			expectErr:  pkg.ErrSHA256NotMatch,
			author:     fileAuthor,
			bucket:     bucket,
		},
		{
			title:      "test upload fail with not match sha256",
			tmp:        tmp2,
			key:        "test/upload.test7",
			expectPath: "",
			expectUri:  "",
			expectErr:  pkg.ErrBucketNotFound,
			author:     fileAuthor,
			bucket:     fmt.Sprintf("bucket-biz-err-%s", time.Now().Format("20060102150405")),
		},
	}
	for _, v := range cases {
		_case := v
		c.Convey(_case.title, t, func() {
			rsp, err := fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
				Key:         _case.key,
				Content:     _case.tmp.content,
				ContentType: _case.tmp.contentType,
				UseVersion:  _case.useVersion,
				Sha256:      _case.tmp.sha256,
				Bucket:      _case.bucket,
			}, _case.author)

			if _case.expectErr != nil {
				c.So(err, c.ShouldNotBeNil)
				c.So(err.Error(), c.ShouldContainSubstring, _case.expectErr.Error())
			} else {
				c.So(err, c.ShouldBeNil)
				c.So(rsp.Uri, c.ShouldEqual, _case.expectUri)
				expectSum, err := sumFileSha256(_case.expectPath)
				c.So(err, c.ShouldBeNil)
				c.So(_case.tmp.sha256, c.ShouldEqual, expectSum)
				c.So(rsp.Uri, c.ShouldEqual, _case.expectUri)
				if _case.useVersion {
					// t.Logf("upload title:%s,version:%s,sha256:%s", _case.title, rsp.Version, _case.tmp.sha256)
					c.So(rsp.Version, c.ShouldNotBeEmpty)
					// versions = append(versions, rsp.Version)
					versions[rsp.Version] = _case.tmp.sha256
				}
			}

		})

	}
	c.Convey("test upload fail", t, func() {
		// c.So(len(versions),c.ShouldEqual,3)
		for v, sha := range versions {
			t.Logf("version:%s,sha256:%s", v, sha)
		}
		fileAuthor2 := fmt.Sprintf("tester-2-%s", time.Now().Format("20060102150405"))

		filePath := filepath.Join(cnf.GetUploadDir(), bucket, fileAuthor2)
		file, err := os.Create(filePath)
		if err != nil {
			t.Error(err)
		}
		defer file.Close()
		rsp, err := fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
			Key:         "test/upload.test5",
			Content:     tmp.content,
			ContentType: tmp.contentType,
			UseVersion:  true,
			Sha256:      tmp.sha256,
			Bucket:      bucket,
		}, fileAuthor2)

		c.So(err, c.ShouldNotBeNil)
		c.So(rsp, c.ShouldBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "mkdir")

		rsp, err = fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
			Key:         "test/.",
			Content:     tmp.content,
			ContentType: tmp.contentType,
			UseVersion:  true,
			Sha256:      tmp.sha256,
			Bucket:      bucket,
		}, fileAuthor)

		c.So(err, c.ShouldNotBeNil)
		c.So(rsp, c.ShouldBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "is a directory")
		fileAuthor3 := fmt.Sprintf("tester-3-%s", time.Now().Format("20060102150405"))
		patch := gomonkey.ApplyFuncReturn((*os.File).Write, 0, fmt.Errorf("write error"))
		defer patch.Reset()
		rsp, err = fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
			Key:         "test/upload.test6",
			Content:     nil,
			ContentType: tmp.contentType,
			UseVersion:  true,
			Sha256:      tmp.sha256,
			Bucket:      bucket,
		}, fileAuthor3)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp, c.ShouldBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "write error")

		patch2 := gomonkey.ApplyFuncReturn(filepath.Rel, "", fmt.Errorf("rel error"))
		defer patch2.Reset()
		rsp, err = fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
			Key:         "test/upload.test6",
			Content:     nil,
			ContentType: tmp.contentType,
			UseVersion:  true,
			Sha256:      tmp.sha256,
			Bucket:      bucket,
		}, fileAuthor3)
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp, c.ShouldBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "rel error")
		patch2.Reset()
	})

}
func testDownload(t *testing.T) {
	fileBiz := newFileBiz()
	versionsKeys := make([]string, 0)
	for v := range versions {
		versionsKeys = append(versionsKeys, v)
	}
	cases := []*FileDownloadTestCase{
		{
			title:      "download success",
			exceptErr:  nil,
			key:        fileAuthor + "/test/upload.test1",
			useVersion: false,
			version:    "",
			author:     fileAuthor,
			sha256:     tmp.sha256,
		},
		{
			title:      "download with version1",
			exceptErr:  nil,
			key:        fileAuthor + "/test/upload.test2",
			useVersion: true,
			version:    versionsKeys[0],
			author:     fileAuthor,
			sha256:     versions[versionsKeys[0]],
		},
		{
			title:      "download with version2",
			exceptErr:  nil,
			key:        fileAuthor + "/test/upload.test2",
			useVersion: true,
			version:    versionsKeys[1],
			author:     fileAuthor,
			sha256:     versions[versionsKeys[1]],
		},
		{
			title:      "download with version3",
			exceptErr:  nil,
			key:        fileAuthor + "/test/upload.test2",
			useVersion: true,
			version:    versionsKeys[2],
			author:     fileAuthor,
			sha256:     versions[versionsKeys[2]],
		},
		{
			title:      "download with latest version",
			exceptErr:  nil,
			key:        fileAuthor + "/test/upload.test2",
			useVersion: true,
			version:    "",
			author:     fileAuthor,
			sha256:     tmp.sha256,
		},

		{
			title:      "download fail with invalidate key",
			exceptErr:  fmt.Errorf("%s", "no such file or directory"),
			key:        fileAuthor + "/test/upload_not_exist",
			useVersion: false,
			version:    "",
			author:     fileAuthor,
			sha256:     tmp.sha256,
		},
		{
			title:      "download fail with version and invalidate key",
			exceptErr:  fmt.Errorf("%s", "file not found"),
			key:        fileAuthor + "/test/upload_not_exist",
			useVersion: true,
			version:    "latest",
			author:     fileAuthor,
			sha256:     tmp.sha256,
		},
	}
	for _, v := range cases {
		_case := v
		c.Convey(_case.title, t, func() {
			rsp, err := fileBiz.Download(context.TODO(), &api.DownloadRequest{
				Key:     _case.key,
				Bucket:  bucket,
				Version: _case.version,
			}, _case.author)
			if _case.exceptErr != nil {
				c.So(err, c.ShouldNotBeNil)
				c.So(err.Error(), c.ShouldContainSubstring, _case.exceptErr.Error())
			} else {
				c.So(err, c.ShouldBeNil)
				c.So(rsp, c.ShouldNotBeNil)

				shaer := sha256.New()
				shaer.Write(rsp)
				// t.Logf("download,title:%s,version:%s,sha256:%s", _case.title, _case.version,hex.EncodeToString(shaer.Sum(nil)))
				c.So(hex.EncodeToString(shaer.Sum(nil)), c.ShouldEqual, _case.sha256)
			}
		})
	}
	c.Convey("test download fail", t, func() {
		_, err := fileBiz.Download(context.TODO(), &api.DownloadRequest{
			Key:    "/" + fileAuthor + "/test/upload.test1",
			Bucket: bucket,
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrInvalidFileKey.Error())

		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env

		}
		config := config.ReadConfig(env)
		cnf := cfg.NewConfig(config)
		filePath := filepath.Join(cnf.GetUploadDir(), bucket, fileAuthor, "test", "upload.test2")
		t.Logf("filepath:%s", filePath)
		reader, err := file.NewFileVersionReader(filePath, "latest")
		c.So(err, c.ShouldBeNil)
		patch := gomonkey.ApplyMethodReturn(reader, "Reader", nil, fmt.Errorf("reader error"))
		defer patch.Reset()
		_, err = fileBiz.Download(context.TODO(), &api.DownloadRequest{
			Key:     fileAuthor + "/test/upload.test2",
			Bucket:  bucket,
			Version: "latest",
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "reader error")
		patch.Reset()
		file, err := file.NewFileVersionReader(filepath.Join(cnf.GetUploadDir(), bucket, fileAuthor, "test", "upload.test2"), "latest")
		c.So(err, c.ShouldBeNil)
		ioReader, err := file.Reader()
		c.So(err, c.ShouldBeNil)
		patch2 := gomonkey.ApplyMethodReturn(ioReader, "Read", 0, fmt.Errorf("error read file"))
		defer patch2.Reset()
		_, err = fileBiz.Download(context.TODO(), &api.DownloadRequest{
			Key:     fileAuthor + "/test/upload.test2",
			Version: "latest",
			Bucket:  bucket,
		}, fileAuthor)

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "error read file")
		patch2.Reset()

		// patch2:=gomonkey.ApplyFuncReturn((*object.File).)
	})
}

func testInitiateUploadFile(t *testing.T) {
	fileBiz := newFileBiz()
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	c.Convey("test init parts upload file success", t, func() {
		rsp, err := fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
			Key: "test/upload.parts.test1",
		})
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		c.So(rsp.UploadId, c.ShouldNotBeEmpty)
		partsPath := filepath.Join(cnf.GetUploadDir(), rsp.UploadId, "parts")
		_, err = os.Stat(partsPath)
		c.So(err, c.ShouldBeNil)
		uploadId = rsp.UploadId

	})
	c.Convey("test init parts upload file fail", t, func() {
		_, err := fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
			Key: "/test/upload.parts.test1",
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrInvalidFileKey.Error())

		patch := gomonkey.ApplyFuncReturn(os.MkdirAll, fmt.Errorf("mkdir error"))
		defer patch.Reset()
		_, err = fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
			Key: "test/upload.parts.test1",
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "mkdir error")

	})

}
func uploadParts(src string, uploadId string, key string, t *testing.T) error {
	fileBiz := newFileBiz()

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("Failed to get file info: %w", err)
	}
	partSize := int64(2 * 1024 * 1024)
	partCount := math.Ceil(float64(info.Size()) / float64(partSize))
	batchSize := 0
	wg := &sync.WaitGroup{}
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Failed to open file: %w", err)
	}
	defer file.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sha := sha256.New()
	for i := 0; i < int(partCount); i++ {
		buffer := make([]byte, partSize)
		n, err := file.Read(buffer)
		sha.Write(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("Failed to read file: %w", err)
		}
		if n == 0 {
			break
		}
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, uploadId string, partNumber int, data []byte) {
			defer wg.Done()
			shaer := sha256.New()
			shaer.Write(data)
			c.Convey(fmt.Sprintf("upload %s for %d part", uploadId, partNumber), t, func() {
				rsp, err := fileBiz.UploadMultipartFileFile(ctx, &api.UploadMultipartFileRequest{
					Key:        key,
					UploadId:   uploadId,
					PartNumber: int64(partNumber),
					Content:    data,
					Sha256:     hex.EncodeToString(shaer.Sum(nil)),
				})
				c.So(err, c.ShouldBeNil)
				if err != nil || rsp == nil {
					cancel()
					return
				}
			})

		}(ctx, wg, uploadId, i+1, buffer)
		batchSize++
		if batchSize == 10 {
			wg.Wait()
			batchSize = 0
		}

	}
	if batchSize > 0 {
		wg.Wait()
	}
	return nil
}
func testUploadMultipartFileFile(t *testing.T) {
	bigTmpFile, err := generateRandomFile(1024 * 1024 * 12)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(bigTmpFile.path)
	bigFileSha256 = bigTmpFile.sha256
	fileBiz := newFileBiz()

	c.Convey("test upload parts file success", t, func() {

		err = uploadParts(bigTmpFile.path, uploadId, "test/upload.parts.test1", t)
		c.So(err, c.ShouldBeNil)
	})

	c.Convey("test upload parts file fail", t, func() {
		// err = uploadParts(bigTmpFile.path, uploadId, "test/upload.parts.test1", t)
		_, err := fileBiz.UploadMultipartFileFile(context.TODO(), &api.UploadMultipartFileRequest{
			Key:        "test/upload.parts.test1",
			UploadId:   "123455678098",
			PartNumber: 1,
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrUploadNotInitiate.Error())

		_, err = fileBiz.UploadMultipartFileFile(context.TODO(), &api.UploadMultipartFileRequest{
			Key:        "test/upload.parts.test1",
			UploadId:   "123455678098",
			PartNumber: 0,
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrPartNumberMissing.Error())

		_, err = fileBiz.UploadMultipartFileFile(context.TODO(), &api.UploadMultipartFileRequest{
			Key:        "test/upload.parts.test1",
			UploadId:   "",
			PartNumber: 1,
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrUploadIdMissing.Error())

		tmpFile2, err := generateRandomFile(1024 * 1024 * 2)
		if err != nil {
			t.Error(err)
		}
		defer os.Remove(tmpFile2.path)
		rsp, err := fileBiz.InitiateUploadFile(context.Background(), &api.InitiateMultipartUploadRequest{Key: "test/upload.parts.test2"})
		c.So(err, c.ShouldBeNil)
		_, err = fileBiz.UploadMultipartFileFile(context.TODO(), &api.UploadMultipartFileRequest{
			Key:        "test/upload.parts.test1",
			UploadId:   rsp.UploadId,
			PartNumber: 1,
			Content:    tmpFile2.content,
			Sha256:     "12233333333er23ed",
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrSHA256NotMatch.Error())

		patch := gomonkey.ApplyFuncReturn(os.Create, nil, fmt.Errorf("file create error"))
		defer patch.Reset()
		_, err = fileBiz.UploadMultipartFileFile(context.TODO(), &api.UploadMultipartFileRequest{
			Key:        "test/upload.parts.test1",
			UploadId:   rsp.UploadId,
			PartNumber: 1,
			Content:    tmpFile2.content,
			Sha256:     tmpFile2.sha256,
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "file create error")
		patch.Reset()

		patch2 := gomonkey.ApplyFuncReturn((*os.File).Write, 0, fmt.Errorf("write error"))
		defer patch2.Reset()

		_, err = fileBiz.UploadMultipartFileFile(context.TODO(), &api.UploadMultipartFileRequest{
			Key:        "test/upload.parts.test1",
			UploadId:   rsp.UploadId,
			PartNumber: 1,
			Content:    tmpFile2.content,
			Sha256:     tmpFile2.sha256,
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "write error")
		patch2.Reset()

	})

}
func testAbortMultipartUpload(t *testing.T) {
	fileBiz := newFileBiz()
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	c.Convey("test abort parts file success", t, func() {
		rsp, err := fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
			Key: "test/upload.parts.test_deleted",
		})
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)

		bigTmpFile, err := generateRandomFile(1024 * 1024 * 12)
		if err != nil {
			t.Error(err)
		}
		err = uploadParts(bigTmpFile.path, rsp.UploadId, "test/upload.parts.test_deleted", t)
		c.So(err, c.ShouldBeNil)

		rsp2, err := fileBiz.AbortMultipartUpload(context.TODO(), &api.AbortMultipartUploadRequest{UploadId: rsp.UploadId})
		c.So(err, c.ShouldBeNil)
		c.So(rsp2, c.ShouldNotBeNil)
		_, err = os.Stat(filepath.Join(cnf.GetUploadDir(), rsp.UploadId, "parts"))
		c.So(err, c.ShouldNotBeNil)

	})
	c.Convey("test abort parts file fail", t, func() {
		_, err := fileBiz.AbortMultipartUpload(context.TODO(), &api.AbortMultipartUploadRequest{UploadId: "fwefwfccccc"})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrUploadIdNotFound.Error())
	})
}
func testCompleteMultipartUploadFile(t *testing.T) {
	fileBiz := newFileBiz()
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	c.Convey("test complete parts file success", t, func() {
		rsp, err := fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
			Key:        "test/upload.parts.test1",
			UploadId:   uploadId,
			UseVersion: true,
			Sha256:     bigFileSha256,
			Bucket:     bucket,
			Engine: api.FileEngine_FILE_ENGINE_LOCAL.String(),
		}, fileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		c.So(rsp.Uri, c.ShouldNotBeEmpty)
		uploadRootDir := cnf.GetUploadDir()

		filePath := filepath.Join(uploadRootDir, bucket, rsp.Uri)
		sha, err := sumFileSha256(filePath)

		if err != nil {
			t.Error(err)
		}
		c.So(sha, c.ShouldEqual, bigFileSha256)

	})
	c.Convey("test complete parts file fail", t, func() {

		_, err := fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
			Key:      "test/upload.parts.test2",
			UploadId: "123455678098",
			Bucket:   bucket,
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrUploadIdNotFound.Error())

		_, err = fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
			Key:      "test/upload.parts.test2",
			UploadId: "123455678098",
		}, "")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrIdentityMissing.Error())

		_, err = fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
			Key:      "/test/upload.parts.test2",
			UploadId: "123455678098",
			Bucket:   bucket,
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrInvalidFileKey.Error())

		patch := gomonkey.ApplyFuncReturn(os.Lstat, nil, fmt.Errorf("lstat error"))
		defer patch.Reset()
		rsp, _ := fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
			Key: "test/upload.parts.test2",
		})
		uploadId2 := rsp.UploadId
		_, err = fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
			Key:        "test/upload.parts.test2",
			UploadId:   uploadId2,
			UseVersion: true,
			Bucket:     bucket,
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "lstat error")
		patch.Reset()

		cases := []struct {
			patch  interface{}
			output []interface{}
			err    error
		}{
			{
				patch:  os.MkdirAll,
				output: []interface{}{fmt.Errorf("mkdir error")},
				err:    fmt.Errorf("mkdir error"),
			}, {
				patch:  os.Create,
				output: []interface{}{nil, fmt.Errorf("create error")},
				err:    fmt.Errorf("create error"),
			}, {
				patch:  os.Open,
				output: []interface{}{nil, fmt.Errorf("open error")},
				err:    fmt.Errorf("open error"),
			},
			{
				patch:  io.Copy,
				output: []interface{}{int64(0), fmt.Errorf("copy error")},
				err:    fmt.Errorf("copy error"),
			},
			{
				patch:  os.RemoveAll,
				output: []interface{}{fmt.Errorf("remove all error")},
				err:    fmt.Errorf("remove all error"),
			}, {
				patch:  filepath.Rel,
				output: []interface{}{"", fmt.Errorf("rel error")},
				err:    fmt.Errorf("rel error"),
			}, {
				patch:  git.PlainInit,
				output: []interface{}{nil, fmt.Errorf("init error")},
				err:    fmt.Errorf("init error"),
			},
		}

		for _, cases := range cases {
			rsp, _ = fileBiz.InitiateUploadFile(context.TODO(), &api.InitiateMultipartUploadRequest{
				Key: "test/upload.parts.test2",
			})
			uploadId3 := rsp.UploadId
			bigTmpFile, _ := generateRandomFile(1024 * 1024 * 2)
			defer os.Remove(bigTmpFile.path)
			_ = uploadParts(bigTmpFile.path, uploadId3, "test/upload.parts.test2", t)
			os.Remove(bigTmpFile.path)
			patch2 := gomonkey.ApplyFuncReturn(cases.patch, cases.output...)
			defer patch2.Reset()
			_, err = fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
				Key:        "test/upload.parts.test2",
				UploadId:   uploadId3,
				UseVersion: true,
				Bucket:     bucket,
			}, fileAuthor)
			patch2.Reset()

			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, cases.err.Error())

		}
	})

}
func testFileMeta(t *testing.T) {
	fileBiz := newFileBiz()
	c.Convey("test file metadata", t, func() {
		meta, err := fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "",
			Bucket:  bucket,
			Engine: api.FileEngine_FILE_ENGINE_LOCAL.String(),
		}, fileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(meta, c.ShouldNotBeNil)
		c.So(meta.Size, c.ShouldEqual, 1024*1024*12)
		c.So(meta.Sha256, c.ShouldEqual, bigFileSha256)

		meta, err = fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "latest",
			Bucket:  bucket,
			Engine: api.FileEngine_FILE_ENGINE_LOCAL.String(),
		}, fileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(meta, c.ShouldNotBeNil)
		c.So(meta.Size, c.ShouldEqual, 1024*1024*12)
		c.So(meta.Sha256, c.ShouldEqual, bigFileSha256)
	})
	c.Convey("test file metadata fail", t, func() {
		_, err := fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{
			Key:     "",
			Version: "",
			Bucket:  bucket,
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrInvalidFileKey.Error())
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := config.ReadConfig(env)
		cnf := cfg.NewConfig(conf)
		filePath := filepath.Join(cnf.GetUploadDir(), bucket, filepath.Dir(fileAuthor+"/test/upload.parts.test1"))
		filePath = filepath.Join(filePath, filepath.Base(fileAuthor+"/test/upload.parts.test1"))

		patch := gomonkey.ApplyFuncReturn(file.NewFileVersionReader, nil, fmt.Errorf("file NewFileVersionReader error"))
		defer patch.Reset()
		_, err = fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "latest",
			Bucket:  bucket,
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "file NewFileVersionReader error")
		patch.Reset()

		reader, err := file.NewFileVersionReader(filePath, "latest")
		c.So(err, c.ShouldBeNil)
		patch2 := gomonkey.ApplyMethodReturn(reader, "Reader", nil, fmt.Errorf("file Reader error"))
		defer patch2.Reset()
		_, err = fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "latest",
			Bucket:  bucket,
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "file Reader error")
		patch2.Reset()

		patch3 := gomonkey.ApplyFuncReturn(io.Copy, int64(0), fmt.Errorf("io.copy error"))
		defer patch3.Reset()
		_, err = fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{Key: fileAuthor + "/test/upload.parts.test1",
			Version: "latest", Bucket: bucket}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "io.copy error")
		patch3.Reset()

		patch4:=gomonkey.ApplyFuncReturn(tiga.MySQLDao.First,fmt.Errorf("get error"))
		defer patch4.Reset()
		_, err = fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "latest",
			Bucket:  bucket,
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get error")
		patch4.Reset()

	})
}
func testVersionReader(t *testing.T) {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	c.Convey("test version reader", t, func() {
		root := cnf.GetUploadDir()
		path := filepath.Join(root, bucket, fileAuthor, "test", "upload.parts.test1")
		reader, err := file.NewFileVersionReader(path, "")
		c.So(err, c.ShouldBeNil)
		c.So(reader, c.ShouldNotBeNil)
		c.So(reader.Author(), c.ShouldEqual, fileAuthor)
		c.So(reader.Name(), c.ShouldEqual, path)
		c.So(reader.ModifyTime(), c.ShouldBeGreaterThan, 0)
		fileReader, err := reader.Reader()
		c.So(err, c.ShouldBeNil)
		c.So(fileReader, c.ShouldNotBeNil)

	})
	c.Convey("test version reader fail", t, func() {
		fileBiz := newFileBiz()
		_, err := fileBiz.Version(context.Background(), bucket, "/test/version.test", fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrInvalidFileKey.Error())
		patch := gomonkey.ApplyFuncReturn(file.NewFileVersionReader, nil, fmt.Errorf("file NewFileVersionReader error"))
		defer patch.Reset()
		_, err = fileBiz.Version(context.Background(), bucket, fileAuthor+"/test/upload.parts.test1", fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "file NewFileVersionReader error")
		patch.Reset()
		patch2 := gomonkey.ApplyFuncReturn(git.PlainOpen, nil, fmt.Errorf("git PlainOpen error"))
		defer patch2.Reset()
		_, err = file.NewFileVersionReader(fileAuthor+"/test/upload.parts.test1", "latest")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "git PlainOpen error")
		patch2.Reset()

		patch3 := gomonkey.ApplyFuncReturn((*git.Repository).Head, nil, fmt.Errorf("head error"))
		defer patch3.Reset()
		root := cnf.GetUploadDir()
		path := filepath.Join(root, bucket, fileAuthor, "test", "upload.parts.test2")
		_, err = file.NewFileVersionReader(path, "latest")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "head error")
		patch3.Reset()

		filePath := filepath.Join(cnf.GetUploadDir(), bucket, filepath.Dir(fileAuthor+"/test/upload.parts.test1"))
		filePath = filepath.Join(filePath, filepath.Base(fileAuthor+"/test/upload.parts.test1"))
		reader, err := file.NewFileVersionReader(filePath, "latest")
		c.So(err, c.ShouldBeNil)
		cases := []struct {
			patch  interface{}
			output []interface{}
			err    error
		}{
			{
				patch:  (*git.Repository).CommitObject,
				output: []interface{}{nil, fmt.Errorf("commit object error")},
				err:    fmt.Errorf("commit object error"),
			},
			{
				patch:  (*object.Commit).Tree,
				output: []interface{}{nil, fmt.Errorf("tree error")},
				err:    fmt.Errorf("tree error"),
			},
			{
				patch:  (*object.Tree).File,
				output: []interface{}{nil, fmt.Errorf("file error")},
				err:    fmt.Errorf("file error"),
			},
			{
				patch: io.CopyN,
				output: []interface{}{
					int64(0), fmt.Errorf("copy error"),
				},
				err: fmt.Errorf("copy error"),
			},
			{
				patch:  io.ReadFull,
				output: []interface{}{0, fmt.Errorf("read full error")},
				err:    fmt.Errorf("read full error"),
			},
		}
		for _, cases := range cases {
			patch4 := gomonkey.ApplyFuncReturn(cases.patch, cases.output...)
			defer patch4.Reset()
			_, err := reader.ReadAt(make([]byte, 1024), 512)
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, cases.err.Error())
			patch4.Reset()
		}
		// _, err = fileBiz.Version(context.Background(), fileAuthor+"/test/upload.parts.test1", fileAuthor)
	})
}
func testDownloadRange(t *testing.T) {
	fileBiz := newFileBiz()

	fileSize := 1024 * 1024 * 12

	c.Convey("test download range parts file success", t, func() {
		partSize := int64(2 * 1024 * 1024)
		partCount := math.Ceil(float64(fileSize) / float64(partSize))
		shaer := sha256.New()
		for i := 0; i < int(partCount); i++ {
			rangeStartAt := int64(i) * partSize
			rangeEndAt := rangeStartAt + partSize - 1
			if rangeEndAt > int64(fileSize) {
				rangeEndAt = int64(fileSize) - 1
			}
			data, _, err := fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
				Key:     fileAuthor + "/test/upload.parts.test1",
				Version: "latest",
				Bucket:  bucket,
			}, rangeStartAt, rangeEndAt, fileAuthor)
			c.So(err, c.ShouldBeNil)
			shaer.Write(data)
		}
		b := shaer.Sum(nil)
		c.So(hex.EncodeToString(b), c.ShouldEqual, bigFileSha256)

		shaer2 := sha256.New()
		data, _, err := fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "latest",
			Bucket:  bucket,
		}, 0, 0, fileAuthor)
		c.So(err, c.ShouldBeNil)
		shaer2.Write(data)
		c.So(hex.EncodeToString(shaer2.Sum(nil)), c.ShouldEqual, bigFileSha256)

		shaer3 := sha256.New()
		data, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "",
			Bucket:  bucket,
		}, 0, 0, fileAuthor)
		c.So(err, c.ShouldBeNil)
		shaer3.Write(data)
		c.So(hex.EncodeToString(shaer3.Sum(nil)), c.ShouldEqual, bigFileSha256)
	})
	c.Convey("test download range parts file fail", t, func() {
		_, _, err := fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:     fileAuthor + "/test/upload.parts.test40",
			Version: "lasted",
			Bucket:  bucket,
		}, 0, 1024, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, plumbing.ErrObjectNotFound.Error())

		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key: "/test/upload.parts.test1",
		}, 0, 1024, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrInvalidFileKey.Error())

		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:     "test/upload.parts.test1",
			Version: "latest",
			Bucket:  bucket,
		}, 1024, 1, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrInvalidRange.Error())

		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "latest",
			Bucket:  bucket,
		}, 1024, 0, fileAuthor)
		c.So(err, c.ShouldBeNil)
		patch := gomonkey.ApplyFuncReturn(file.NewFileVersionReader, nil, git.ErrRepositoryNotExists)
		defer patch.Reset()
		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:     "test/upload.parts.test1",
			Version: "latest",
			Bucket:  bucket,
		}, 0, 1024, fileAuthor)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, git.ErrRepositoryNotExists.Error())
		patch.Reset()
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := config.ReadConfig(env)
		cnf := cfg.NewConfig(conf)
		filePath := filepath.Join(cnf.GetUploadDir(), bucket, filepath.Dir(fileAuthor+"/test/upload.parts.test1"))
		filePath = filepath.Join(filePath, filepath.Base(fileAuthor+"/test/upload.parts.test1"))
		reader, err := file.NewFileVersionReader(filePath, "latest")
		c.So(err, c.ShouldBeNil)
		patch2 := gomonkey.ApplyMethodReturn(reader, "ReadAt", 0, fmt.Errorf("file readAt error"))
		defer patch2.Reset()

		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "latest",
			Bucket:  bucket,
		}, 0, 1024, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "file readAt error")
		patch2.Reset()

	})
}
func testList(t *testing.T) {
	fileBiz := newFileBiz()
	c.Convey("test list file success", t, func() {
		rsp, err := fileBiz.List(context.Background(), &api.ListFilesRequest{
			Bucket:   bucket,
			Page:     1,
			PageSize: 20,
		}, fileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		c.So(rsp, c.ShouldNotBeEmpty)
	})
	c.Convey("test list file fail", t, func() {
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.Pagination, fmt.Errorf("file not found"))
		defer patch.Reset()
		_, err := fileBiz.List(context.Background(), &api.ListFilesRequest{
			Bucket:   fmt.Sprintf("test-%s", time.Now().Format("20060102150405")),
			Page:     -1,
			PageSize: 20,
		}, fileAuthor)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "file not found")
	})

}
func testDelete(t *testing.T) {
	fileBiz := newFileBiz()
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)

	c.Convey("test delete file fail", t, func() {
		patch := gomonkey.ApplyFuncReturn(file.NewFileReader, nil, fmt.Errorf("file not found"))
		defer patch.Reset()
		_, err := fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: fileAuthor + "/test/upload.parts.deleted", Bucket: bucket}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "file not found")
		patch.Reset()

		_, err = fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: fileAuthor + "/test/upload.parts.deleted2", Bucket: bucket}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")

		_, err = fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: "/" + fileAuthor + "/test/upload.parts.deleted2", Bucket: bucket}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, pkg.ErrInvalidFileKey.Error())

		patch2 := gomonkey.ApplyFuncReturn(os.RemoveAll, fmt.Errorf("remove all error"))
		defer patch2.Reset()
		_, err = fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: fileAuthor + "/test/upload.parts.test1", Bucket: bucket}, fileAuthor)
		patch2.Reset()

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "remove all error")

	})
	c.Convey("test delete file success", t, func() {
		rsp, err := fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: fileAuthor + "/test/upload.parts.test1", Bucket: bucket}, fileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		_, err = os.Stat(filepath.Join(cnf.GetUploadDir(), fileAuthor, "test", "upload.parts.test1"))
		c.So(err, c.ShouldNotBeNil)
		_, err = os.Stat(filepath.Join(cnf.GetUploadDir(), "parts", fileAuthor+"/test/upload.parts.test1"))
		c.So(err, c.ShouldNotBeNil)

	})

}

func testUploadVersionCommitErr(t *testing.T) {
	fileBiz := newFileBiz()
	c.Convey("test upload version commit error", t, func() {
		var err error
		tmp2, err := generateRandomFile(1024 * 1024 * 1)
		if err != nil {
			t.Error(err)
		}
		defer os.Remove(tmp2.path)
		cases := []struct {
			patch  interface{}
			output []interface{}
			err    error
		}{
			{
				patch:  git.PlainInit,
				output: []interface{}{nil, fmt.Errorf("commit error")},
				err:    fmt.Errorf("commit error"),
			},
			{
				patch:  (*git.Repository).Worktree,
				output: []interface{}{nil, fmt.Errorf("worktree error")},
				err:    fmt.Errorf("worktree error"),
			},
			{
				patch:  (*git.Worktree).Add,
				output: []interface{}{nil, fmt.Errorf("add error")},
				err:    fmt.Errorf("add error"),
			},
			{
				patch:  (*git.Worktree).Commit,
				output: []interface{}{nil, fmt.Errorf("commit error")},
				err:    fmt.Errorf("commit error"),
			},
			{
				patch:  (*git.Repository).CommitObject,
				output: []interface{}{nil, fmt.Errorf("commit object error")},
				err:    fmt.Errorf("commit object error"),
			},
		}
		fileAuthor2 := fmt.Sprintf("tester-fail-%s", time.Now().Format("20060102150405"))
		for _, caseV := range cases {
			patch := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch.Reset()
			_, err := fileBiz.Upload(context.TODO(), &api.UploadFileRequest{
				Key:         "test/upload_version.test1",
				Content:     tmp2.content,
				ContentType: tmp2.contentType,
				UseVersion:  true,
				Sha256:      tmp2.sha256,
				Bucket:      bucket,
			}, fileAuthor2)
			patch.Reset()

			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())
		}

		patch := gomonkey.ApplyFuncReturn(git.PlainInit, nil, git.ErrRepositoryAlreadyExists)
		patch = patch.ApplyFuncReturn(git.PlainOpen, nil, fmt.Errorf("open error"))
		defer patch.Reset()
		f := &api.UploadFileRequest{
			Key:         "test/upload_version.test1",
			Content:     tmp2.content,
			ContentType: tmp2.contentType,
			UseVersion:  true,
			Sha256:      tmp2.sha256,
			Bucket:      bucket,
		}
		_, err = fileBiz.Upload(context.TODO(), f, fileAuthor2)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "open error")
		patch.Reset()

		patch2 := gomonkey.ApplyFuncReturn((*git.Worktree).Commit, nil, git.ErrEmptyCommit)
		patch2 = patch2.ApplyFuncSeq((*git.Repository).Head, []gomonkey.OutputCell{{Times: 2, Values: []interface{}{plumbing.NewHashReference(plumbing.ReferenceName("test"), plumbing.NewHash("xxxxxxxxxx")), nil}}, {Values: []interface{}{nil, fmt.Errorf("get head ref error")}}})
		defer patch2.Reset()
		_, err = fileBiz.Upload(context.TODO(), f, fileAuthor2)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get head ref error")
		t.Logf("test upload version commit error end,%s", err.Error())
		patch2.Reset()

	})
}
func TestFile(t *testing.T) {
	t.Run("test make bucket", testMkBucket)
	t.Run("test upload", testPutFile)
	t.Run("test upload version commit error", testUploadVersionCommitErr)
	t.Run("test download", testDownload)
	t.Run("test initiate upload file", testInitiateUploadFile)
	t.Run("test upload parts file", testUploadMultipartFileFile)
	t.Run("test complete parts file", testCompleteMultipartUploadFile)
	t.Run("test file metadata", testFileMeta)
	t.Run("test version reader", testVersionReader)
	t.Run("test abort parts file", testAbortMultipartUpload)
	t.Run("test download range parts file", testDownloadRange)
	t.Run("test list file", testList)
	t.Run("test delete file", testDelete)

}
