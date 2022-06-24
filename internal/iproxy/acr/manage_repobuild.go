package acr

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// CancelRepoBuild 取消仓库构建任务。
func (c *Client) CancelRepoBuild() (response *responses.CommonResponse, err error) {
	return
}

// CreateRepoBuildRule 创建一个仓库的构建规则。
func (c *Client) CreateRepoBuildRule(repoNamespace, repoName, dockerfileLocation, tag string) (data BuildRule, err error) {
	request := c.NewRequest()
	request.PathPattern = fmt.Sprintf("/repos/%s/%s/rules", repoNamespace, repoName)
	request.Method = requests.PUT
	content := fmt.Sprintf(
		`{
			"BuildRule": {
				"PushType": "GIT_BRANCH",
				"PushName": "master",
				"DockerfileLocation": "%s",
				"DockerfileName": "Dockerfile",
				"Tag": "%s"
			}
		}`, dockerfileLocation, tag,
	)
	request.SetContent([]byte(content))
	response, err := c.ProcessCommonRequest(request)

	err = decodeData(response.GetHttpContentBytes(), &data)
	return
}

// DeleteRepoBuildRule 删除一个仓库的构建规则。
func (c *Client) DeleteRepoBuildRule() (response *responses.CommonResponse, err error) {
	return
}

// GetRepoBuildList 查询仓库构建记录。
func (c *Client) GetRepoBuildList(repoNamespace, repoName string) (data Builds, err error) {
	request := c.NewRequest()
	request.PathPattern = fmt.Sprintf("/repos/%s/%s/build", repoNamespace, repoName)
	response, err := c.ProcessCommonRequest(request)

	err = decodeData(response.GetHttpContentBytes(), &data)
	return
}

// GetRepoBuildRuleList 查询仓库构建规则列表。
func (c *Client) GetRepoBuildRuleList(repoNamespace, repoName string) (data BuildRules, err error) {
	request := c.NewRequest()
	request.PathPattern = fmt.Sprintf("/repos/%s/%s/rules", repoNamespace, repoName)
	response, err := c.ProcessCommonRequest(request)

	err = decodeData(response.GetHttpContentBytes(), &data)
	return
}

// GetRepoBuildStatus 查询仓库中构建任务的状态。
func (c *Client) GetRepoBuildStatus(repoNamespace, repoName, buildID string) (data Build, err error) {
	request := c.NewRequest()
	request.PathPattern = fmt.Sprintf("/repos/%s/%s/build/%s/status", repoNamespace, repoName, buildID)
	response, err := c.ProcessCommonRequest(request)

	err = decodeData(response.GetHttpContentBytes(), &data)
	return
}

// StartRepoBuildByRule 根据仓库构建规则创建构建任务。
func (c *Client) StartRepoBuildByRule(repoNamespace, repoName string, buildRuleID int) (response *responses.CommonResponse, err error) {
	request := c.NewRequest()
	request.PathPattern = fmt.Sprintf("/repos/%s/%s/rules/%d/build", repoNamespace, repoName, buildRuleID)
	request.Method = requests.PUT
	response, err = c.ProcessCommonRequest(request)
	return
}

// UpdateRepoBuildRule 更新仓库构建规则。
func (c *Client) UpdateRepoBuildRule(repoNamespace, repoName, dockerfileLocation, tag string, buildRule BuildRule) (data BuildRule, err error) {
	request := c.NewRequest()
	request.PathPattern = fmt.Sprintf("/repos/%s/%s/rules/%d", repoNamespace, repoName, buildRule.BuildRuleID)
	request.Method = requests.POST
	content := fmt.Sprintf(
		`{
			"BuildRule": {
				"PushType": "GIT_BRANCH",
				"PushName": "master",
				"DockerfileLocation": "%s",
				"DockerfileName": "Dockerfile",
				"Tag": "%s"
			}
		 }`, dockerfileLocation, tag,
	)
	request.SetContent([]byte(content))
	response, err := c.ProcessCommonRequest(request)

	err = decodeData(response.GetHttpContentBytes(), &data)
	return
}
