package acr

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// GetNamespaceList 查询命名空间列表。
func (c *Client) GetNamespaceList() (response *responses.CommonResponse, err error) {
	request := c.NewRequest()
	request.PathPattern = "/namespace"
	response, err = c.ProcessCommonRequest(request)
	return
}

// GetNamespace 查询指定命名空间的详细信息。
func (c *Client) GetNamespace(namespace string) (response *responses.CommonResponse, err error) {
	request := c.NewRequest()
	request.PathPattern = fmt.Sprintf("/namespace/%s", namespace)
	response, err = c.ProcessCommonRequest(request)
	return
}

// UpdateNamespace 更新命名空间的基本信息。
// AutoCreate true or false
// DefaultVisibility "PUBLIC" or "PRIVATE"
func (c *Client) UpdateNamespace(namespace string, AutoCreate bool, DefaultVisibility string) (response *responses.CommonResponse, err error) {
	request := c.NewRequest()
	request.Method = requests.POST
	content := fmt.Sprintf(
		`{
			"Namespace":{
				"AutoCreate":%v,
				"DefaultVisibility":"%s"
			}
		}`, AutoCreate, DefaultVisibility,
	)
	request.SetContent([]byte(content))
	request.PathPattern = fmt.Sprintf("/namespace/%s", namespace)
	response, err = c.ProcessCommonRequest(request)
	return
}

// CreateNamespace 创建一个新的命名空间。
func (c *Client) CreateNamespace(namespace string) (response *responses.CommonResponse, err error) {
	request := c.NewRequest()
	request.Method = requests.PUT
	content := fmt.Sprintf(
		`{
			"Namespace": {
				"Namespace": "%s",
			}
		}`, namespace,
	)
	request.SetContent([]byte(content))
	request.PathPattern = "/namespace"
	response, err = c.ProcessCommonRequest(request)
	return
}

// DeleteNamespace 删除一个已有命名空间，注意这个操作会将存在于该命名空间下的所有仓库以及所有仓库下的镜像一并删除。
func (c *Client) DeleteNamespace(namespace string) (response *responses.CommonResponse, err error) {
	request := c.NewRequest()
	request.Method = requests.DELETE
	request.PathPattern = fmt.Sprintf("/namespace/%s", namespace)
	response, err = c.ProcessCommonRequest(request)
	return
}
