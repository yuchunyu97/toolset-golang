package main

// Email Sender

// 交叉编译 Windows
// 在 tools/cmd/email-sender 目录下执行
// CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o email-sender-v0.0.2.exe .

import (
	"crypto/tls"
	"fmt"
	"github.com/jordan-wright/email"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	ToolVersion = "v0.0.2"
	ToolAuthor  = "AIslandX <yuchunyu97@gmail.com>"
)

var (
	ConfigFile     = "email-sender-config.xlsx"
	AttachmentPath = "email-sender-attachments"
)

func main() {
	fmt.Printf("欢迎使用群发邮件小工具\nversion %s\nauthor %s\n\n", ToolVersion, ToolAuthor)

	AttachmentPath = filepath.Join("configs", AttachmentPath)
	ConfigFile = filepath.Join("configs", ConfigFile)
	f, err := excelize.OpenFile(ConfigFile)
	if err != nil {
		log.Printf("读取配置文件 %s 失败 %s", ConfigFile, err)
		waitForExit()
		return
	}

	// https://github.com/qax-os/excelize
	// 是否将邮件密送自己
	// 因为通过 SMTP 发送的邮件已发送邮件中不会显示，开启后可以确定邮件是否正确发送
	bccYourself := false
	// 配置字段
	var server, username, password, bccYourselfConfig string
	if server, err = f.GetCellValue("email-config", "A2"); err != nil {
		log.Printf("读取配置 server 失败，"+
			"请将 server 信息写入 email-config Sheet 的 A2 单元格中，如 mail.exchange.com:587 %s", err)
		waitForExit()
		return
	}
	if username, err = f.GetCellValue("email-config", "B2"); err != nil {
		log.Printf("读取配置 email 失败，"+
			"请将 email 信息写入 email-config Sheet 的 B2 单元格中，如 username@exchange.com %s", err)
		waitForExit()
		return
	}
	if password, err = f.GetCellValue("email-config", "C2"); err != nil {
		log.Printf("读取配置 password 失败，"+
			"请将 password 信息写入 email-config Sheet 的 C2 单元格中 %s", err)
		waitForExit()
		return
	}
	if bccYourselfConfig, err = f.GetCellValue("email-config", "D2"); err == nil {
		if bccYourselfConfig == "yes" {
			bccYourself = true
		}
	}

	serverInfo := strings.Split(server, ":")
	if len(serverInfo) != 2 {
		log.Printf("配置 server 格式错误，格式为 serverName:smtpPort，用冒号将服务器名和 SMTP 端口连接起来")
		waitForExit()
		return
	}

	passwordEnc := ""
	for range password {
		passwordEnc += "*"
	}
	log.Printf("配置信息：\n\tServer: %s\n\tUsername(Email): %s\n\tPassword: %s\n\tBCC Yourself: %v\n\n",
		server, username, passwordEnc, bccYourself)

	// 读取待发送的邮件列表
	rows, err := f.GetRows("email-list")
	if err != nil {
		log.Printf("获取待发送的邮件列表失败，请确保 Sheet email-list 存在 %s", err)
		waitForExit()
		return
	}
	var emailToBeSendList []*email.Email
	// 逐行处理
	for idx := range rows {
		if idx > 0 {
			rowNum := idx + 1
			log.Printf("开始处理 Excel email-list Sheet 第 %d 行\n", rowNum)

			var cellEmailTo, cellEmailCc, cellEmailSubject, cellEmailHTML, cellEmailAttachments string
			var cellErr error
			cellEmailTo, cellErr = f.GetCellValue("email-list", fmt.Sprintf("A%d", rowNum))
			cellEmailCc, cellErr = f.GetCellValue("email-list", fmt.Sprintf("B%d", rowNum))
			cellEmailSubject, cellErr = f.GetCellValue("email-list", fmt.Sprintf("C%d", rowNum))
			cellEmailHTML, cellErr = f.GetCellValue("email-list", fmt.Sprintf("D%d", rowNum))
			cellEmailAttachments, cellErr = f.GetCellValue("email-list", fmt.Sprintf("E%d", rowNum))

			if cellErr != nil {
				log.Printf("发生错误，跳过 %s", cellErr)
				continue
			}

			// 去除收件人和抄送人邮件两边的空格
			cellEmailTo = strings.TrimSpace(cellEmailTo)
			cellEmailCc = strings.TrimSpace(cellEmailCc)

			// 初始化邮件
			em := email.NewEmail()
			em.From = username

			// * 收件人邮箱（用英文分号分隔）
			emailTo := strings.Split(cellEmailTo, ";")
			if len(emailTo) == 0 {
				log.Printf("收件人邮箱必填，跳过")
				continue
			}
			for _, v := range emailTo {
				vv := strings.TrimSpace(v)
				if vv != "" {
					em.To = append(em.To, vv)
				}
			}

			// 抄送人邮箱（用英文分号分隔）
			emailCc := strings.Split(cellEmailCc, ";")
			if len(emailCc) > 0 {
				for _, v := range emailCc {
					vv := strings.TrimSpace(v)
					if vv != "" {
						em.Cc = append(em.Cc, vv)
					}
				}
			}

			// 密送自己
			if bccYourself {
				em.Bcc = []string{username}
			}

			// * 邮件主题
			if cellEmailSubject == "" {
				log.Printf("邮件主题必填，跳过")
				continue
			}
			em.Subject = cellEmailSubject

			// * 邮件正文文件名（文件在当前目录下）
			if cellEmailHTML == "" {
				log.Printf("邮件正文文件名必填，跳过")
				continue
			}
			emailBody, err := os.ReadFile(cellEmailHTML)
			if err != nil {
				log.Printf("邮件正文读取失败，跳过 %s", err)
				continue
			}
			em.HTML = emailBody

			// 邮件附件文件名（文件放在 email-sender-attachments 文件夹中，多个文件用英文逗号分隔）
			if cellEmailAttachments != "" {
				emailAttachmentsList := strings.Split(cellEmailAttachments, ",")
				for _, attachFileName := range emailAttachmentsList {
					attachFilePath := filepath.Join(AttachmentPath, attachFileName)
					_, err = em.AttachFile(attachFilePath)
					if err != nil {
						log.Printf("添加邮件附件 %s 失败，跳过 %s", attachFilePath, err)
						break
					}
				}
				// 如果附件没有全部添加成功，则跳过
				if len(emailAttachmentsList) != len(em.Attachments) {
					continue
				}
			}

			// 待发送邮件添加成功
			log.Printf("待发送邮件添加成功")
			emailToBeSendList = append(emailToBeSendList, em)
		}
	}
	log.Printf("--------------------\n\n")

	if len(emailToBeSendList) == 0 {
		log.Printf("待发送邮件列表为空，退出")
		return
	}

	// 打印出待发送的邮件列表，请用户确认是否发送
	fmt.Printf("待发送邮件列表如下：\n\n")
	for idx, emailInfo := range emailToBeSendList {
		fmt.Printf("NO.%d\nSubject: %v\nAttatchments Count: %d\n",
			idx+1, emailInfo.Subject, len(emailInfo.Attachments))

		for _, v := range emailInfo.To {
			fmt.Println("To ->", v)
		}
		for _, v := range emailInfo.Cc {
			fmt.Println("Cc ->", v)
		}
		for _, v := range emailInfo.Bcc {
			fmt.Println("Bcc ->", v)
		}
		fmt.Println()
	}

	var confirmSend string
	fmt.Printf("请确认是否进行发送，发送（Y），不发送（N）：")
	_, _ = fmt.Scanln(&confirmSend)
	if strings.ToLower(confirmSend) != "y" {
		log.Printf("不发送，退出")
		return
	}

	// 开始发送邮件
	// 使用 GOLANG 发送邮件 https://segmentfault.com/a/1190000040310170
	log.Printf("开始发送邮件")

	//设置服务器相关的配置
	auth := LoginAuth(username, password)
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	successSendCount := 0
	for idx, emailInfo := range emailToBeSendList {
		log.Printf("正在发送第 %d 封邮件给 %s ：%s（%d 个附件）\n",
			idx+1, emailInfo.To, emailInfo.Subject, len(emailInfo.Attachments))

		// 防止发送过于频繁被限流，暂停 5s
		time.Sleep(time.Second * 5)

		err = emailInfo.SendWithStartTLS(server, auth, tlsConfig)
		if err != nil {
			log.Printf("发送失败 %s\n", err)
			continue
		}

		log.Printf("发送成功\n")
		successSendCount += 1
	}

	log.Printf("\n\n发送成功 %d 封邮件，失败 %d 封邮件\n\n", successSendCount, len(emailToBeSendList)-successSendCount)

	fmt.Printf("按任意键退出")
	_, _ = fmt.Scanln()
}

func waitForExit() {
	fmt.Printf("按任意键退出")
	_, _ = fmt.Scanln()
}
