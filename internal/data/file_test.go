package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	cfg "github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	api "github.com/begonia-org/go-sdk/api/file/v1"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var fileBucketId = ""
var fileFileId = ""

func testUpsertFile(t *testing.T) {
	c.Convey("Test UpsertFile", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		f := NewFileRepo(conf, gateway.Log)
		snk, _ := tiga.NewSnowflake(1)
		fileFileId = snk.GenerateIDString()
		fk := fmt.Sprintf("test-%s", fileFileId)
		// log.Printf("fileFileId:%s,file id %s",fileFileId,fk)
		file := &api.Files{
			Engine:     fk,
			Bucket:     "test",
			Key:        fk,
			Uid:        fileFileId,
			Owner:      "test",
			CreatedAt:  timestamppb.Now(),
			UpdatedAt:  timestamppb.Now(),
			UpdateMask: &fieldmaskpb.FieldMask{},
		}
		// t.Logf("fileFileId:%s,file id %s",fileFileId,file.Uid)
		updated, err := f.UpsertFile(context.Background(), file)
		c.So(err, c.ShouldBeNil)
		c.So(updated, c.ShouldBeFalse)
		time.Sleep(3 * time.Second)
		file.UpdatedAt = timestamppb.Now()
		file.UpdateMask.Paths = append(file.UpdateMask.Paths, "updated_at")
		updated, err = f.UpsertFile(context.Background(), file)

		c.So(err, c.ShouldBeNil)
		c.So(updated, c.ShouldBeTrue)
	})
}
func testGetFileById(t *testing.T) {
	c.Convey("Test GetFileById", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		f := NewFileRepo(conf, gateway.Log)
		file, err := f.GetFileById(context.Background(), fileFileId)
		c.So(err, c.ShouldBeNil)
		c.So(file, c.ShouldNotBeNil)
		fk:=fmt.Sprintf("test-%s",fileFileId)
		file, err = f.GetFile(context.Background(), fk, "test", fk)
		c.So(err, c.ShouldBeNil)
		c.So(file, c.ShouldNotBeNil)
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.First, fmt.Errorf("get error"))
		defer patch.Reset()
		_, err = f.GetFileById(context.Background(), "")

		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get error")
		_, err = f.GetFile(context.Background(), "test", "test", "test")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "get error")

	})
}
func testUpsertBucket(t *testing.T) {
	c.Convey("Test UpsertBucket", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		f := NewFileRepo(conf, gateway.Log)
		snk, _ := tiga.NewSnowflake(1)
		bk:=fmt.Sprintf("test-%s",snk.GenerateIDString())

		bucket := &api.Buckets{
			Engine:     "test",
			Bucket:     bk,
			Uid:        snk.GenerateIDString(),
			Owner:      "test",
			CreatedAt:  timestamppb.Now(),
			UpdatedAt:  timestamppb.Now(),
			UpdateMask: &fieldmaskpb.FieldMask{},
		}
		fileBucketId = bucket.Uid
		updated, err := f.UpsertBucket(context.Background(), bucket)
		c.So(err, c.ShouldBeNil)
		c.So(updated, c.ShouldBeFalse)
		time.Sleep(3 * time.Second)
		bucket.UpdatedAt = timestamppb.Now()
		bucket.UpdateMask.Paths = append(bucket.UpdateMask.Paths, "updated_at")
		updated, err = f.UpsertBucket(context.Background(), bucket)
		c.So(err, c.ShouldBeNil)
		c.So(updated, c.ShouldBeTrue)
	})
}
func testFileList(t *testing.T) {
	c.Convey("Test file List", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		f := NewFileRepo(conf, gateway.Log)
		files, err := f.List(context.Background(), 1, 10, "test", "test", "test")
		c.So(err, c.ShouldBeNil)
		c.So(files, c.ShouldNotBeNil)

		files, err = f.List(context.Background(), 1, 10, "", "test", "test")
		c.So(err, c.ShouldBeNil)
		c.So(files, c.ShouldNotBeNil)
	})
	c.Convey("Test file List fail", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		f := NewFileRepo(conf, gateway.Log)
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.Pagination, fmt.Errorf("page error"))
		defer patch.Reset()

		files, err := f.List(context.Background(), 1, 10, "", "", "test")
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "page error")
		c.So(files, c.ShouldBeEmpty)
	})
}
func testDeleteFile(t *testing.T) {
	c.Convey("Test DeleteFile", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		f := NewFileRepo(conf, gateway.Log)
		err := f.DelFile(context.Background(), "test", "test", "test")
		c.So(err, c.ShouldBeNil)
	})
}
func testDeleteBucket(t *testing.T) {
	c.Convey("Test DeleteBucket", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		conf := cfg.ReadConfig(env)
		f := NewFileRepo(conf, gateway.Log)
		err := f.DelBucket(context.Background(), fileBucketId)
		c.So(err, c.ShouldBeNil)
	})
}
func TestFile(t *testing.T) {
	t.Run("TestUpsertFile", testUpsertFile)
	t.Run("TestGetFileById", testGetFileById)
	t.Run("TestUpsertBucket", testUpsertBucket)
	t.Run("TestFileList", testFileList)
	t.Run("TestDeleteFile", testDeleteFile)
	t.Run("TestDeleteBucket", testDeleteBucket)
}