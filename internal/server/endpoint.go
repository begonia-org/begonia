package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"

	"github.com/begonia-org/begonia/endpoint"
	"github.com/begonia-org/begonia/internal/biz"
	"github.com/begonia-org/begonia/internal/pkg/config"
	"github.com/begonia-org/begonia/internal/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EndpointManager interface {
	Watch(ctx context.Context, dir string, errChan chan<- error) error
}
type EndpointManagerImpl struct {
	biz    *biz.EndpointUsecase
	config *config.Config
}

func NewEndpointManagerImpl(biz *biz.EndpointUsecase, config *config.Config) EndpointManager {
	return &EndpointManagerImpl{biz: biz, config: config}
}
func (imp *EndpointManagerImpl) addEndpoints(reg endpoint.EndpointRegister, endpoint string) error {
	GlobalMutex.Lock()
	defer GlobalMutex.Unlock()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	return reg.RegisterAll(context.Background(), GlobalMux, endpoint, opts)
}
func (imp *EndpointManagerImpl) createEndpointRegister(pluginPath string) (endpoint.EndpointRegister, error) {
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, err
	}
	newEndpointSymbol, err := p.Lookup("NewEndpointRegisters")
	if err != nil {
		return nil, err
	}
	newRegisterFunc := newEndpointSymbol.(func() endpoint.EndpointRegister)
	newEndpointRegisters := newRegisterFunc()
	return newEndpointRegisters, nil
}
func (imp *EndpointManagerImpl) getPluginSoFile(dir string) (string, error) {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		return "", err
	}
	if fileInfo == nil {
		return "", fmt.Errorf("not found dir")
	}
	if !fileInfo.IsDir() {
		return "", fmt.Errorf("not dir")
	}
	files, err := filepath.Glob("*.so")
	if err != nil {
		return "", err
	}
	for _, file := range files {
		return file, nil
	}
	return "", fmt.Errorf("not found so file")
}
func (imp *EndpointManagerImpl) Watch(ctx context.Context, dir string, errChan chan<- error) error {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watch.Close()
	err = watch.Add(dir)
	if err != nil {

		return err
	}
	logger.Logger.Info("watching dir: ", dir)
	for {
		select {
		case ev := <-watch.Events:
			{
				if ev.Op&fsnotify.Create == fsnotify.Create {
					logger.Logger.Infof("创建文件 : %s, %s", ev.Name, ev.String())
					// 创建文件时，加载插件
					filename := filepath.Base(ev.Name)
					filename = strings.TrimSuffix(filename, ".so")
					endpoint, err := imp.biz.GetEndpoint(context.Background(), filename)
					if err != nil {
						errChan <- err
						logger.Logger.Errorf("get endpoint error: %s", err.Error())
						continue
					}
					filepath, err := imp.getPluginSoFile(ev.Name)
					if err != nil {
						errChan <- err
						logger.Logger.Errorf("get plugin so file error: %s", err.Error())
						continue
					}
					endpointRegister, err := imp.createEndpointRegister(filepath)
					if err != nil {
						errChan <- err
						logger.Logger.Errorf("create endpoint register error: %s", err.Error())
						continue
					}
					err = imp.addEndpoints(endpointRegister, endpoint.Endpoint)
					if err != nil {
						errChan <- err
						logger.Logger.Errorf("add endpoint error: %s", err.Error())
						continue
					}
				}

			}
		case err := <-watch.Errors:
			{
				logger.Logger.Error(err)
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}
