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
	"github.com/begonia-org/begonia/internal/biz/file"
	api "github.com/begonia-org/go-sdk/api/file/v1"
	"github.com/go-git/go-git/v5/plumbing"

	cfg "github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/errors"
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

func newFileBiz() *file.FileUsecase {
	env := "dev"
	if begonia.Env != "" {
		env = begonia.Env
	}
	config := config.ReadConfig(env)
	cnf := cfg.NewConfig(config)
	return file.NewFileUsecase(cnf)
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
	fileAuthor = fmt.Sprintf("tester-%s", time.Now().Format("20060102150405"))
	fileSha256 = tmp.sha256
	cases := []fileUploadTestCase{
		{
			title:      "test upload without version",
			tmp:        tmp,
			useVersion: false,
			key:        "test/upload.test1",
			expectPath: filepath.Join(cnf.GetUploadDir(), fileAuthor, "test", "upload.test1"),
			expectUri:  fileAuthor + "/test/upload.test1",
			expectErr:  nil,
			author:     fileAuthor,
		},
		{
			title:      "test upload with version",
			tmp:        tmp,
			useVersion: true,
			key:        "test/upload.test2",
			expectPath: filepath.Join(cnf.GetUploadDir(), fileAuthor, "test", "upload.test2"),
			expectUri:  fileAuthor + "/test/upload.test2",
			expectErr:  nil,
			author:     fileAuthor,
		},
		{
			title:      "test upload with version1",
			tmp:        tmp,
			useVersion: true,
			key:        "test/upload.test2",
			expectPath: filepath.Join(cnf.GetUploadDir(), fileAuthor, "test", "upload.test2"),
			expectUri:  fileAuthor + "/test/upload.test2",
			expectErr:  nil,
			author:     fileAuthor,
		},
		{
			title:      "test upload with version2",
			tmp:        tmp3,
			useVersion: true,
			key:        "test/upload.test2",
			expectPath: filepath.Join(cnf.GetUploadDir(), fileAuthor, "test", "upload.test2"),
			expectUri:  fileAuthor + "/test/upload.test2",
			expectErr:  nil,
			author:     fileAuthor,
		},
		{
			title:      "test upload with version3",
			tmp:        tmp,
			useVersion: true,
			key:        "test/upload.test2",
			expectPath: filepath.Join(cnf.GetUploadDir(), fileAuthor, "test", "upload.test2"),
			expectUri:  fileAuthor + "/test/upload.test2",
			expectErr:  nil,
			author:     fileAuthor,
		},
		// {
		// 	title:      "test upload with version4",
		// 	tmp:        tmp,
		// 	useVersion: true,
		// 	key:        "test/upload.test2",
		// 	expectPath: filepath.Join(cnf.GetUploadDir(), fileAuthor, "test", "upload.test2"),
		// 	expectUri:  fileAuthor + "/test/upload.test2",
		// 	expectErr:  nil,
		// 	author:     fileAuthor,
		// },
		{
			title:      "test upload with invalid key",
			tmp:        tmp,
			key:        "/test/upload.test3",
			expectPath: "",
			expectUri:  "",
			expectErr:  errors.ErrInvalidFileKey,
			author:     fileAuthor,
		},
		{
			title:      "test upload with invalid author",
			tmp:        tmp,
			key:        "test/upload.test4",
			expectPath: "",
			expectUri:  "",
			expectErr:  errors.ErrIdentityMissing,
			author:     "",
		},
		{
			title:      "test upload fail with not match sha256",
			tmp:        tmp2,
			key:        "test/upload.test7",
			expectPath: "",
			expectUri:  "",
			expectErr:  errors.ErrSHA256NotMatch,
			author:     fileAuthor,
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

		filePath := filepath.Join(cnf.GetUploadDir(), fileAuthor2)
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
		}, fileAuthor3)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(rsp, c.ShouldBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "write error")
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
		// {
		// 	title:      "download fail with invalidate author",
		// 	exceptErr:  errors.ErrIdentityMissing,
		// 	key:        fileAuthor + "/test/upload.test1",
		// 	useVersion: true,
		// 	version:    "",
		// 	author:     "",
		// 	sha256:     tmp.sha256,
		// },
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
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrInvalidFileKey.Error())

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
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrUploadNotInitiate.Error())

		_, err = fileBiz.UploadMultipartFileFile(context.TODO(), &api.UploadMultipartFileRequest{
			Key:        "test/upload.parts.test1",
			UploadId:   "123455678098",
			PartNumber: 0,
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrPartNumberMissing.Error())

		_, err = fileBiz.UploadMultipartFileFile(context.TODO(), &api.UploadMultipartFileRequest{
			Key:        "test/upload.parts.test1",
			UploadId:   "",
			PartNumber: 1,
		})
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrUploadIdMissing.Error())

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
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrSHA256NotMatch.Error())

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
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrUploadIdNotFound.Error())
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
		}, fileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		c.So(rsp.Uri, c.ShouldNotBeEmpty)
		uploadRootDir := cnf.GetUploadDir()

		filePath := filepath.Join(uploadRootDir, rsp.Uri)
		sha, err := sumFileSha256(filePath)

		if err != nil {
			t.Error(err)
		}
		c.So(sha, c.ShouldEqual, bigFileSha256)

	})
	c.Convey("test complete parts file fail", t, func() {
		_, err := fileBiz.CompleteMultipartUploadFile(context.TODO(), &api.CompleteMultipartUploadRequest{
			Key:      "test/upload.parts.test1",
			UploadId: "123455678098",
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrUploadIdNotFound.Error())

	})
}
func testFileMeta(t *testing.T) {
	fileBiz := newFileBiz()
	c.Convey("test file metadata", t, func() {
		meta, err := fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{
			Key:     fileAuthor + "/test/upload.parts.test1",
			Version: "",
		}, fileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(meta, c.ShouldNotBeNil)
		c.So(meta.Size, c.ShouldEqual, 1024*1024*12)
		c.So(meta.Sha256, c.ShouldEqual, bigFileSha256)
	})
	c.Convey("test file metadata fail", t, func() {
		// _, err := fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{
		// 	Key:     fileAuthor + "/test/upload.parts.test1",
		// 	Version: "",
		// }, "")
		// c.So(err, c.ShouldNotBeNil)
		// c.So(err.Error(), c.ShouldContainSubstring, errors.ErrIdentityMissing.Error())

		_, err := fileBiz.Metadata(context.Background(), &api.FileMetadataRequest{
			Key:     "",
			Version: "",
		}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrInvalidFileKey.Error())
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
		path := filepath.Join(root, fileAuthor, "test", "upload.parts.test1")
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
			}, rangeStartAt, rangeEndAt, fileAuthor)
			c.So(err, c.ShouldBeNil)
			shaer.Write(data)
		}
		b := shaer.Sum(nil)
		c.So(hex.EncodeToString(b), c.ShouldEqual, bigFileSha256)
	})
	c.Convey("test download range parts file fail", t, func() {
		_, _, err := fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key:     fileAuthor + "/test/upload.parts.test40",
			Version: "lasted",
		}, 0, 1024, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, plumbing.ErrObjectNotFound.Error())

		_, _, err = fileBiz.DownloadForRange(context.Background(), &api.DownloadRequest{
			Key: "/test/upload.parts.test1",
		}, 0, 1024, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, errors.ErrInvalidFileKey.Error())
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
	c.Convey("test delete file success", t, func() {
		rsp, err := fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: fileAuthor + "/test/upload.parts.test1"}, fileAuthor)
		c.So(err, c.ShouldBeNil)
		c.So(rsp, c.ShouldNotBeNil)
		_, err = os.Stat(filepath.Join(cnf.GetUploadDir(), fileAuthor, "test", "upload.parts.test1"))
		c.So(err, c.ShouldNotBeNil)
		_, err = os.Stat(filepath.Join(cnf.GetUploadDir(), "parts", fileAuthor+"/test/upload.parts.test1"))
		c.So(err, c.ShouldNotBeNil)

	})
	c.Convey("test delete file fail", t, func() {
		patch := gomonkey.ApplyFuncReturn(file.NewFileReader, nil, fmt.Errorf("file not found"))
		defer patch.Reset()
		_, err := fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: fileAuthor + "/test/upload.parts.deleted"}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "file not found")
		patch.Reset()

		_, err = fileBiz.Delete(context.Background(), &api.DeleteRequest{Key: fileAuthor+"/test/upload.parts.deleted2"}, fileAuthor)
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "not found")

	})

}
func TestFile(t *testing.T) {
	t.Run("test upload", testPutFile)
	t.Run("test download", testDownload)
	t.Run("test initiate upload file", testInitiateUploadFile)
	t.Run("test upload parts file", testUploadMultipartFileFile)
	t.Run("test complete parts file", testCompleteMultipartUploadFile)
	t.Run("test file metadata", testFileMeta)
	t.Run("test version reader", testVersionReader)
	t.Run("test abort parts file", testAbortMultipartUpload)
	t.Run("test download range parts file", testDownloadRange)
	t.Run("test delete file", testDelete)

}
