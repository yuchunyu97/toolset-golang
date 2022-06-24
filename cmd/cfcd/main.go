package main

// CFCD, Compress folders in the current directory

// 交叉编译 Windows
// CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o cfcd-v0.0.2.exe main.go

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
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

type NeedZipInfo struct {
	Name string
	Path string
}

func main() {
	fmt.Printf("欢迎使用压缩小工具\nversion %s\nauthor %s\n\n", ToolVersion, ToolAuthor)

	// 从控制台获取输入的需要被压缩的目录
	// 目录中所有的文件夹会单独被压缩成 zip 包
	var inputZipDir string
	fmt.Printf("请输入需要压缩的目录：")
	if _, err := fmt.Scanln(&inputZipDir); err != nil {
		log.Printf("输入错误 %s", err)
		waitForExit()
		return
	}

	// 判断输入是否为目录
	inputZipDirInfo, err := os.Stat(inputZipDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("需要压缩的目录 %s 不存在", inputZipDir)
			waitForExit()
			return
		} else {
			log.Printf("获取需要压缩的目录 %s 出错 %s", inputZipDir, err)
			waitForExit()
			return
		}
	}
	if !inputZipDirInfo.IsDir() {
		log.Printf("%s 不是目录", inputZipDir)
		waitForExit()
		return
	}
	fmt.Println()

	// 输入的值校验通过，获取当前目录下需要压缩的目录列表
	var needZipList []NeedZipInfo
	// 读取目录下文件
	files, err := ioutil.ReadDir(inputZipDir)
	if err != nil {
		log.Printf("获取文件列表出错 %s", err)
		waitForExit()
		return
	}
	for _, file := range files {
		if file.IsDir() {
			needZipList = append(needZipList, NeedZipInfo{
				Name: file.Name(),
				Path: filepath.Join(inputZipDir, file.Name()),
			})
		}
	}

	// 当前程序执行目录
	pwd, _ := os.Getwd()

	// 在当前目录下创建结果输出文件夹，命名格式 cfcd_result_20211217175612
	resultDirName := fmt.Sprintf("cfcd_result_%s_%s",
		inputZipDirInfo.Name(), time.Now().Format("20060102150405"))
	resultDirPath := filepath.Join(pwd, resultDirName)
	if err = os.Mkdir(resultDirPath, os.ModePerm); err != nil {
		log.Printf("创建结果输出文件夹 %s 出错 %s", resultDirPath, err)
		waitForExit()
		return
	}

	for _, needZipInfo := range needZipList {
		compressDirPath := needZipInfo.Path
		zipFileName := fmt.Sprintf("%s.zip", needZipInfo.Name)
		zipFilePath := filepath.Join(resultDirPath, zipFileName)

		log.Printf("开始压缩目录 %s", compressDirPath)
		if err = Zip(zipFilePath, compressDirPath); err != nil {
			log.Printf("压缩 %s 失败：%s", compressDirPath, err)
		} else {
			log.Printf("压缩成功 %s\n\n", zipFilePath)
		}
	}

	fmt.Printf("按任意键退出")
	_, _ = fmt.Scanln()
}

func Zip(dst, src string) (err error) {
	// https://learnku.com/articles/23434/golang-learning-notes-five-archivezip-to-achieve-compression-and-decompression
	// https://stackoverflow.com/questions/49057032/recursively-zipping-a-directory-in-golang

	// 创建准备写入的文件
	fw, err := os.Create(dst)
	defer func() { _ = fw.Close() }()
	if err != nil {
		return err
	}

	// 通过 fw 来创建 zip.Write
	zw := zip.NewWriter(fw)
	defer func() {
		// 检测一下是否成功关闭
		if err := zw.Close(); err != nil {
			log.Println(err)
			waitForExit()
			return
		}
	}()

	// 下面来将文件写入 zw ，因为有可能会有很多个目录及文件，所以递归处理
	return filepath.Walk(src, func(path string, fi os.FileInfo, errBack error) (err error) {
		if errBack != nil {
			return errBack
		}

		// 通过文件信息，创建 zip 的文件信息
		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return
		}

		// 替换文件信息中的文件名
		fh.Name = strings.TrimPrefix(path, src)

		// 这步开始没有加，会发现解压的时候说它不是个目录
		if fi.IsDir() {
			fh.Name += "/"
		}

		// 写入文件信息，并返回一个 Write 结构
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return
		}

		// 检测，如果不是标准文件就只写入头信息，不写入文件数据到 w
		// 如目录，也没有数据需要写
		if !fh.Mode().IsRegular() {
			return nil
		}

		// 打开要压缩的文件
		fr, err := os.Open(path)
		defer func() { _ = fr.Close() }()
		if err != nil {
			return
		}

		// 将打开的文件 Copy 到 w
		_, err = io.Copy(w, fr)
		if err != nil {
			return
		}
		// 输出压缩的内容
		log.Printf("成功压缩文件： %s\n", fh.Name)

		return nil
	})
}

func waitForExit() {
	fmt.Printf("按任意键退出")
	_, _ = fmt.Scanln()
}
