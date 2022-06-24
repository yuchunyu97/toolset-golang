// Package code Git 实现库
// 文档地址：https://pkg.go.dev/github.com/go-git/go-git/v5
package code

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Code Code
type Code struct {
	GitPath  string
	CodePath string
	Auth     *http.BasicAuth
	repo     *git.Repository
}

// Init 拉取 Git 代码，存入临时文件夹
func (c *Code) Init() (err error) {
	if c.GitPath == "" {
		err = errors.New("Empty git path")
		return
	}

	dir, err := ioutil.TempDir("", "iproxy-code")
	if err != nil {
		return
	}
	c.CodePath = dir

	c.repo, err = git.PlainClone(c.CodePath, false, &git.CloneOptions{
		URL:  c.GitPath,
		Auth: c.Auth,
	})
	if err != nil {
		return
	}
	return
}

// Remove 移除临时文件夹
func (c *Code) Remove() {
	os.RemoveAll(c.CodePath)
}

// AddFile 增加文件
func (c *Code) AddFile(fileSubPath string, name string, content string) (err error) {
	filePath := path.Join(c.CodePath, fileSubPath)
	os.MkdirAll(filePath, os.ModePerm)
	fileNamePath := path.Join(filePath, name)
	file, _ := os.OpenFile(fileNamePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer file.Close()
	file.WriteString(content)
	// 将文件加入 Git 追踪
	// 获取工作树
	w, err := c.repo.Worktree()
	if err != nil {
		return fmt.Errorf("init worktree: %s", err)
	}
	// 添加文件
	_, err = w.Add(path.Join(fileSubPath, name))
	if err != nil {
		return fmt.Errorf("add file to worktree: %s", err)
	}
	return
}

// Submit 提交 Git 仓库
func (c *Code) Submit(msg string) (err error) {
	// 获取工作树
	w, err := c.repo.Worktree()
	if err != nil {
		return fmt.Errorf("init worktree: %s", err)
	}
	// 创建 commit 数据
	commit, err := w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "iproxy",
			Email: "yuchunyu97@gmail.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("generate commit info: %s", err)
	}
	// 提交 commit
	_, err = c.repo.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("create commit: %s", err)
	}
	// push
	err = c.repo.Push(&git.PushOptions{
		Auth: c.Auth,
	})
	if err != nil {
		return fmt.Errorf("push: %s", err)
	}
	return
}

// PrepareCode PrepareCode
func PrepareCode(gitPath, gitUsername, gitPassword string) (mycode *Code, err error) {
	mycode = &Code{
		GitPath: gitPath,
		Auth: &http.BasicAuth{
			Username: gitUsername,
			Password: gitPassword,
		},
	}
	if err = mycode.Init(); err != nil {
		return
	}
	return
}
