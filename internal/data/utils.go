package data

import (
	"fmt"

	"github.com/begonia-org/begonia/internal/pkg/config"
)


func getServiceKey(config *config.Config, id string) string {
	prefix := config.GetEndpointsPrefix()
	return fmt.Sprintf("%s/service/%s", prefix, id)
}
func getServiceNameKey(config *config.Config, name string) string {
	prefix := config.GetEndpointsPrefix()
	return fmt.Sprintf("%s/service_name/%s", prefix, name)
}
func getDetailsKey(config *config.Config, id string) string {
	prefix := config.GetEndpointsPrefix()
	return fmt.Sprintf("%s/details/%s", prefix, id)
}
func getTagsKey(config *config.Config, tag, id string) string {
	prefix := config.GetEndpointsPrefix()
	return fmt.Sprintf("%s/tags/%s/%s", prefix, tag, id)

}
