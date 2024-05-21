package data

// func TestCreateInBatches(t *testing.T) {
// 	c.Convey("TestCreateInBatches", t, func() {
// 		env := "dev"
// 		if begonia.Env != "" {
// 			env = begonia.Env
// 		}
// 		conf := cfg.ReadConfig(env)
// 		repo := NewDataRepo(conf, gateway.Log)
// 		repo.db.AutoMigrate(&example.ExampleTable{})
// 		snk, _ := tiga.NewSnowflake(1)
// 		models := []*example.ExampleTable{
// 			{
// 				Uid: snk.GenerateIDString(),
// 				CreatedAt: tiga.Time(time.Now()),
// 			},
// 		}
// 	})
// }
