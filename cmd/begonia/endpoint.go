package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	goloadbalancer "github.com/begonia-org/go-loadbalancer"
	api "github.com/begonia-org/go-sdk/api/app/v1"
	endpoint "github.com/begonia-org/go-sdk/api/endpoint/v1"
	"github.com/begonia-org/go-sdk/client"
)

var accessKey string
var secret string
var addr string

func readInitAPP(env string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	path := filepath.Join(homeDir, ".begonia")
	path = filepath.Join(path, fmt.Sprintf("admin-app.%s.json", env))
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf(err.Error())
		return

	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	app := &api.Apps{}
	err = decoder.Decode(app)
	if err != nil {
		log.Fatalf(err.Error())
		return

	}
	accessKey = app.AccessKey
	secret = app.Secret
	path2 := filepath.Join(homeDir, ".begonia")
	path2 = filepath.Join(path2, "gateway.json")
	gwFile, err := os.Open(path2)
	if err != nil {
		log.Fatalf(err.Error())
		return

	}
	defer gwFile.Close()
	gwDecoder := json.NewDecoder(gwFile)
	gw := &struct {
		Addr string `json:"addr"`
	}{}
	err = gwDecoder.Decode(gw)
	if err != nil {
		log.Fatalf(fmt.Sprintf("read gateway file error:%s", err.Error()))
		return

	}
	addr = gw.Addr

}
func RegisterEndpoint(env,name string, endpoints []string, pbFile string, opts ...client.EndpointOption) {
	readInitAPP(env)
	pb, err := os.ReadFile(pbFile)
	if err != nil {
		panic(err)

	}
	apiClient := client.NewEndpointAPI(addr, accessKey, secret)
	meta := make([]*endpoint.EndpointMeta, 0)
	for i, v := range endpoints {
		meta = append(meta, &endpoint.EndpointMeta{
			Addr:   v,
			Weight: int32(i),
		})
	}
	endpoint := &endpoint.EndpointSrvConfig{
		DescriptorSet: pb,
		Name:          name,
		ServiceName:   name,
		Description:   name,
		Balance:       string(goloadbalancer.RRBalanceType),
		Endpoints:     meta,
		Tags:          make([]string, 0),
	}
	for _, opt := range opts {
		opt(endpoint)

	}
	rsp, err := apiClient.PostEndpointConfig(context.Background(), endpoint)
	if err != nil {
		log.Fatalf(err.Error())
		panic(err.Error())
	}
	log.Printf("#####################Add Endpoint Success#####################")
	log.Printf("#####################ID:%s####################################", rsp.Id)
}
func UpdateEndpoint(env,id string, mask []string, opts ...client.EndpointOption) {
	readInitAPP(env)
	apiClient := client.NewEndpointAPI(addr, accessKey, secret)
	log.Printf("#####################Update Endpoint###########################")
	patch := &endpoint.EndpointSrvUpdateRequest{}
	patch.UniqueKey = id
	for _, opt := range opts {
		opt(patch)
	}
	_, err := apiClient.PatchEndpointConfig(context.Background(), patch)
	if err != nil {
		log.Fatalf(err.Error())
		panic(err.Error())
	}
	log.Printf("#####################Update Endpoint %s Success#####################", id)

}
func DeleteEndpoint(env,id string) {
	readInitAPP(env)
	apiClient := client.NewEndpointAPI(addr, accessKey, secret)
	log.Printf("#####################Delete Endpoint:%s#####################", id)
	_, err := apiClient.DeleteEndpointConfig(context.Background(), id)
	if err != nil {
		log.Printf("delete err %v ", err)
		panic(err.Error())
	}
	log.Printf("#####################Delete Endpoint Success#####################")
}
