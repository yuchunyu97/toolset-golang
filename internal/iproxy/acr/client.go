// Package acr 阿里云容器镜像服务
// API 文档地址：https://help.aliyun.com/document_detail/72377.html
package acr

import (
	"encoding/json"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// Client 客户端
type Client struct {
	sdk.Client
	region string
}

// EndpointMap 地域
var endpointMap map[string]string

func init() {
	if endpointMap == nil {
		endpointMap = map[string]string{
			"cn-qingdao":     "cr.cn-qingdao.aliyuncs.com",
			"cn-beijing":     "cr.cn-beijing.aliyuncs.com",
			"cn-zhangjiakou": "cr.cn-zhangjiakou.aliyuncs.com",
			"cn-huhehaote":   "cr.cn-huhehaote.aliyuncs.com",
			"cn-hangzhou":    "cr.cn-hangzhou.aliyuncs.com",
			"cn-shanghai":    "cr.cn-shanghai.aliyuncs.com",
			"cn-shenzhen":    "cr.cn-shenzhen.aliyuncs.com",
			"cn-hongkong":    "cr.cn-hongkong.aliyuncs.com",
			"ap-northeast-1": "cr.ap-northeast-1.aliyuncs.com",
			"ap-southeast-1": "cr.ap-southeast-1.aliyuncs.com",
			"ap-southeast-2": "cr.ap-southeast-2.aliyuncs.com",
			"ap-southeast-3": "cr.ap-southeast-3.aliyuncs.com",
			"ap-southeast-5": "cr.ap-southeast-5.aliyuncs.com",
			"ap-south-1":     "cr.ap-south-1.aliyuncs.com",
			"us-east-1":      "cr.us-east-1.aliyuncs.com",
			"us-west-1":      "cr.us-west-1.aliyuncs.com",
			"me-east-1":      "cr.me-east-1.aliyuncs.com",
			"eu-central-1":   "cr.eu-central-1.aliyuncs.com",
		}
	}
}

// NewClient 使用 accesskey 创建一个客户端
func NewClient(regionID, accessKeyID, accessKeySecret string) (client *Client, err error) {
	client = &Client{
		region: regionID,
	}
	err = client.InitWithAccessKey(regionID, accessKeyID, accessKeySecret)
	return
}

// NewRequest 创建一个请求体
func (c *Client) NewRequest() (request *requests.CommonRequest) {
	request = requests.NewCommonRequest()
	domain := endpointMap[c.region]
	request.SetDomain(domain)
	request.Version = "2016-06-07"
	request.SetContentType("JSON")
	return
}

// GetRegionList 例子
// 查询区域列表。
func (c *Client) GetRegionList() (response *responses.CommonResponse, err error) {
	request := c.NewRequest()
	request.PathPattern = "/regions"
	response, err = c.ProcessCommonRequest(request)
	return
}

func decodeData(b []byte, i interface{}) (err error) {
	var rawdata map[string]interface{}
	err = json.Unmarshal(b, &rawdata)
	if err != nil {
		return
	}
	data, err := json.Marshal(rawdata["data"])
	if err != nil {
		return
	}
	err = json.Unmarshal(data, i)
	if err != nil {
		return
	}
	return
}
