package begonia

import "os"

var (
	Version   string
	BuildTime string
	Commit    string
	Env       string
)

func init() {
	env := os.Getenv("BEGONIA_ENV")
	if env != "" {
		Env = env
	}
}
