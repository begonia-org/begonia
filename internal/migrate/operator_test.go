package migrate_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/internal"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/migrate"
	c "github.com/smartystreets/goconvey/convey"
	"github.com/spark-lence/tiga"
)

func TestOperator(t *testing.T) {
	c.Convey("TestOperator", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)

		operator := internal.InitOperatorApp(cnf)
		err := operator.Init()
		c.So(err, c.ShouldBeNil)

		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.First, nil)
		patch = patch.ApplyFuncReturn(tiga.MySQLDao.Create, nil)
		defer patch.Reset()
		operator = internal.InitOperatorApp(cnf)
		err = operator.Init()
		c.So(err, c.ShouldBeNil)
	})
	c.Convey("TestOperator fail", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)

		operator := internal.InitOperatorApp(cnf)
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.AutoMigrate, fmt.Errorf("migration failed"))
		defer patch.Reset()
		err := operator.Init()
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "migration failed")

		patch1 := gomonkey.ApplyFuncReturn(tiga.MySQLDao.First, fmt.Errorf("first failed"))
		defer patch1.Reset()
		operator = internal.InitOperatorApp(cnf)
		err = operator.Init()
		patch1.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "first failed")

		patch2 := gomonkey.ApplyFuncReturn(tiga.MySQLDao.First, nil)
		patch2 = patch2.ApplyFuncReturn(tiga.NewSnowflake, nil, fmt.Errorf("snowflake failed"))
		defer patch2.Reset()
		operator = internal.InitOperatorApp(cnf)
		err = operator.Init()
		patch2.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "snowflake failed")

		patch3 := gomonkey.ApplyFuncReturn(tiga.MySQLDao.First, nil)
		patch3 = patch3.ApplyFuncReturn(tiga.EncryptStructAES, fmt.Errorf("encrypt failed"))
		defer patch3.Reset()
		operator = internal.InitOperatorApp(cnf)
		err = operator.Init()
		patch3.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "encrypt failed")
	})
}

func TestAppOperatorFail(t *testing.T) {
	c.Convey("TestAppOperator fail", t, func() {
		env := "dev"
		if begonia.Env != "" {
			env = begonia.Env
		}
		cnf := config.ReadConfig(env)
		mysql := tiga.NewMySQLDao(cnf)
		operator := migrate.NewAPPOperator(mysql)
		cases := []struct {
			patch  interface{}
			output []interface{}
			err    error
		}{

			{
				patch:  tiga.NewSnowflake,
				output: []interface{}{nil, fmt.Errorf("snowflake failed")},
				err:    fmt.Errorf("snowflake failed"),
			},
			{
				patch:  biz.GenerateAppAccessKey,
				output: []interface{}{"", fmt.Errorf("accessKey failed")},
				err:    fmt.Errorf("accessKey failed"),
			},
			{
				patch:  os.UserHomeDir,
				output: []interface{}{"", fmt.Errorf("homeDir failed")},
				err:    fmt.Errorf("homeDir failed"),
			},
			{
				patch:  os.MkdirAll,
				output: []interface{}{fmt.Errorf("mkdir failed")},
				err:    fmt.Errorf("mkdir failed"),
			},
			{
				patch:  os.Create,
				output: []interface{}{nil, fmt.Errorf("create failed")},
				err:    fmt.Errorf("create failed"),
			},
			{
				patch:  (*json.Encoder).Encode,
				output: []interface{}{fmt.Errorf("encode failed")},
				err:    fmt.Errorf("encode failed"),
			},
		}
		patch := gomonkey.ApplyFuncReturn(tiga.MySQLDao.First, fmt.Errorf("first failed"))
		defer patch.Reset()
		err := operator.InitAdminAPP("test",env)
		patch.Reset()
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "first failed")

		patch1 := gomonkey.ApplyFuncReturn(tiga.MySQLDao.First, nil)
		patch1 = patch1.ApplyFuncReturn(tiga.MySQLDao.Create, fmt.Errorf("create failed"))
		defer patch1.Reset()
		for _, caseV := range cases {
			patch2 := gomonkey.ApplyFuncReturn(caseV.patch, caseV.output...)
			defer patch2.Reset()
			err := operator.InitAdminAPP("test",env)
			patch2.Reset()
			c.So(err, c.ShouldNotBeNil)
			c.So(err.Error(), c.ShouldContainSubstring, caseV.err.Error())

		}
		patch1.Reset()

	})
}
