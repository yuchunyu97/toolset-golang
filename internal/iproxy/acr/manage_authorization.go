package acr

import (
	"encoding/json"
	"fmt"
	"time"
)

// GetAuthorizationToken 例子
// 返回用于登录 Registry 的临时账号和临时密码，
// 临时密码的有效时间为1小时，
// 若使用 STS 方式请求时，临时密码的有效时间等同于本次请求 STS Token 的有效时间。
func (c *Client) GetAuthorizationToken() (auth Authorization, err error) {
	request := c.NewRequest()
	request.PathPattern = "/tokens"
	response, err := c.ProcessCommonRequest(request)
	if err != nil {
		return
	}

	var data map[string]map[string]interface{}
	err = json.Unmarshal(response.GetHttpContentBytes(), &data)
	if err != nil {
		return
	}
	auth.UserName = data["data"]["tempUserName"].(string)
	auth.PassWord = data["data"]["authorizationToken"].(string)
	expireDate := int64(data["data"]["expireDate"].(float64)) / 1000
	timeLayout := "2006-01-02 15:04:05"
	auth.Expire = time.Unix(expireDate, 0).Format(timeLayout)
	return
}

// Authorization Authorization
type Authorization struct {
	UserName string
	PassWord string
	Expire   string
}

func (a Authorization) String() string {
	return fmt.Sprintf(`
Use temporary login credentials to pull the image.

UserName: %[1]s
Password: %[2]s

Example:
  docker login --username=%[1]s registry.cn-qingdao.aliyuncs.com
  Password: %[2]s

Note: The temporary certificate expiration time is %[3]s
	`, a.UserName, a.PassWord, a.Expire)
}
