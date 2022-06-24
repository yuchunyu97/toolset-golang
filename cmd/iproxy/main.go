package main

import (
	"crypto/md5"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/yuchunyu97/toolset-golang/internal/iproxy/acr"
	"github.com/yuchunyu97/toolset-golang/internal/iproxy/code"
	"log"
	"os/exec"
	"strings"
	"time"
)

var (
	doInit     bool
	doDownload bool
	imageName  string

	regionId        string
	accessKeyId     string
	accessKeySecret string
	gitPath         string
	gitUsername     string
	gitPassword     string
	repoNamespace   string
	repoName        string
)

func init() {
	pflag.BoolVarP(&doInit, "init", "i", false, "Init the config to file.")
	pflag.BoolVarP(&doDownload, "download", "d", false, "Download the container image to the local.")
	pflag.StringVarP(&imageName, "name", "n", "", "Name of container image to be downloaded.")

	pflag.Parse()

	// 默认配置
	viper.SetDefault("iproxy.regionId", "<region-id>")
	viper.SetDefault("iproxy.accessKeyId", "<access-key-id>")
	viper.SetDefault("iproxy.accessKeySecret", "<access-key-secret>")
	viper.SetDefault("iproxy.gitPath", "<git-path>")
	viper.SetDefault("iproxy.gitUsername", "<git-username>")
	viper.SetDefault("iproxy.gitPassword", "<gir-password>")
	viper.SetDefault("iproxy.repoNamespace", "<repo-namespace>")
	viper.SetDefault("iproxy.repoName", "<repo-name>")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.toolset")
}

func main() {
	// 初始化配置到文件
	if doInit {
		err := viper.WriteConfig()
		if err != nil {
			log.Println("Please confirm that the config file already exists.")
			log.Println("Otherwise, please create it manually:")
			log.Println(">> mkdir -p $HOME/.toolset && touch $HOME/.toolset/config.yaml")
			log.Fatalf("init config error: %s", err)
		}
		log.Println("init config success at $HOME/.toolset/config.yaml")
		return
	}

	// 加载配置文件
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Please confirm that the config file already exists.")
		log.Println("Use `iproxy -i` to init the config to file, then modify to the appropriate value.")
		log.Fatalf("fatal error config file: %s", err)
	}

	regionId = viper.GetString("iproxy.regionId")
	accessKeyId = viper.GetString("iproxy.accessKeyId")
	accessKeySecret = viper.GetString("iproxy.accessKeySecret")
	gitPath = viper.GetString("iproxy.gitPath")
	gitUsername = viper.GetString("iproxy.gitUsername")
	gitPassword = viper.GetString("iproxy.gitPassword")
	repoNamespace = viper.GetString("iproxy.repoNamespace")
	repoName = viper.GetString("iproxy.repoName")

	// 容器镜像的名称不能为空
	if imageName == "" {
		log.Fatalf("name cannot be empty, use -n to specify")
	}

	imagePath := strings.Replace(imageName, ":", "/", -1)
	imagePath = strings.Replace(imagePath, "@", "/", -1)
	if len(imagePath) > 128 {
		imagePath = imagePath[:128] // imagePath 长度限制 128
	}
	m := md5.Sum([]byte(imageName))
	imageMD5 := fmt.Sprintf("%x", m)

	filename := "Dockerfile"
	content := fmt.Sprintf(`# Date: %s
# MD5: %s
FROM %s
	`, time.Now().Format("2006-01-02 15:04:05"), imageMD5, imageName)

	// =======更新代码=======
	log.Println("正在更新代码")
	repo, err := code.PrepareCode(gitPath, gitUsername, gitPassword) // 下载代码库
	if err != nil {
		log.Fatalln(err)
	}
	if err = repo.AddFile(imagePath, filename, content); err != nil { // 添加文件
		log.Fatalln(err)
	}
	if err = repo.Submit(imageName); err != nil { // 提交 Git
		log.Fatalln(err)
	}
	repo.Remove() // 删除临时目录
	log.Println("更新代码成功")

	// =======构建镜像=======
	if err = pull(imagePath, imageMD5); err != nil {
		log.Fatalln(err)
	}
}

func pull(imagePath, imageMD5 string) (err error) {
	// 初始化一个阿里云客户端
	client, err := acr.NewClient(regionId, accessKeyId, accessKeySecret)
	if err != nil {
		return fmt.Errorf("init acr client error: %s", err)
	}

	// 获取现有构建规则列表，并找到可用的构建规则 ID，更新构建规则
	log.Println("更新构建规则")
	var buildRuleID int
	repoBuildRules, err := client.GetRepoBuildRuleList(repoNamespace, repoName)
	if err != nil {
		return fmt.Errorf("get repo build rule list error: %s", err)
	}
	if len(repoBuildRules.BuildRules) == 0 {
		// 新建构建规则
		repoNewBuildRules, err := client.CreateRepoBuildRule(repoNamespace, repoName, imagePath, imageMD5)
		if err != nil {
			return fmt.Errorf("create repo build rule error: %s", err)
		}
		buildRuleID = repoNewBuildRules.BuildRuleID
	} else {
		// 更新构建规则
		repoReBuildRules, err := client.UpdateRepoBuildRule(repoNamespace, repoName, imagePath, imageMD5, repoBuildRules.BuildRules[0])
		if err != nil {
			return fmt.Errorf("update repo build rule error: %s", err)
		}
		buildRuleID = repoReBuildRules.BuildRuleID
	}

	// 触发构建
	log.Println("准备拉取镜像")
	_, err = client.StartRepoBuildByRule(repoNamespace, repoName, buildRuleID)
	if err != nil {
		return fmt.Errorf("start repo build by rule error: %s", err)
	}

	// 通过列表获取构建的 buildID
	var buildID string
	resBuilds, err := client.GetRepoBuildList(repoNamespace, repoName)
	if err != nil {
		return fmt.Errorf("get repo build list error: %s", err)
	}
	// 数组按时间倒序
	length := len(resBuilds.Builds)
	for i := 0; i < length/2; i++ {
		temp := resBuilds.Builds[length-1-i]
		resBuilds.Builds[length-1-i] = resBuilds.Builds[i]
		resBuilds.Builds[i] = temp
	}
	for _, v := range resBuilds.Builds {
		if v.Image.Tag == imageMD5 {
			buildID = v.BuildID
		}
	}

	// 轮询是否构建完成
	currentStatus := ""
	for {
		resBuild, err := client.GetRepoBuildStatus(repoNamespace, repoName, buildID)
		if err != nil {
			return fmt.Errorf("get repo build status error: %s", err)
		}

		if currentStatus == resBuild.BuildStatus {
			fmt.Printf(".")
		} else {
			currentStatus = resBuild.BuildStatus
			fmt.Printf("\n%s", currentStatus)
		}

		if currentStatus == "SUCCESS" {
			fmt.Printf("\n\n")
			log.Println("拉取镜像成功")
			break
		} else if currentStatus == "PENDING" || currentStatus == "BUILDING" {
			time.Sleep(time.Second * 5)
			continue
		} else {
			// status == "FAILED" || status == "CANCELED"
			return fmt.Errorf("镜像拉取失败 Build ID: %s", buildID)
		}
	}

	// 输出拉取镜像的命令
	fmt.Printf(`
使用如下命令拉取镜像：

docker pull registry.%[5]s.aliyuncs.com/%[3]s/%[4]s:%[1]s
docker tag registry.%[5]s.aliyuncs.com/%[3]s/%[4]s:%[1]s %[2]s
docker rmi registry.%[5]s.aliyuncs.com/%[3]s/%[4]s:%[1]s

`, imageMD5, imageName, repoNamespace, repoName, regionId)

	// 下载镜像
	if doDownload {
		imageOrigin := fmt.Sprintf("registry.%s.aliyuncs.com/%s/%s:%s", regionId, repoNamespace, repoName, imageMD5)

		err := RunCommand("docker", "pull", imageOrigin)
		if err != nil {
			return fmt.Errorf("镜像 pull 失败：%s", err)
		}

		err = RunCommand("docker", "tag", imageOrigin, imageName)
		if err != nil {
			return fmt.Errorf("镜像 tag 失败：%s", err)
		}

		err = RunCommand("docker", "rmi", imageOrigin)
		if err != nil {
			return fmt.Errorf("镜像 rmi 失败：%s\n", err)
		}
	}

	return nil
}

func RunCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	fmt.Println(cmd.String())
	// 命令的错误输出和标准输出都连接到同一个管道
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	// 从管道中实时获取输出并打印到终端
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil
}
