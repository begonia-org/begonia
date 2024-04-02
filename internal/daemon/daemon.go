package daemon

import (
	"context"

	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewDaemonImpl)

type Daemon interface {
	Start(ctx context.Context)
}

type DaemonImpl struct {
	config   *config.Config
	operator *biz.DataOperatorUsecase
}

func NewDaemonImpl(config *config.Config, operator *biz.DataOperatorUsecase) Daemon {
	return &DaemonImpl{
		config:   config,
		operator: operator,
	}
}

// Start starts the daemon
//
// It will start the operator to do some operations
// It is a blocking function
func (d *DaemonImpl) Start(ctx context.Context) {
	go d.operator.Do(ctx)
}
