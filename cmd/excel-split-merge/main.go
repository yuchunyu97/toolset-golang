package main

// Excel Split Merge

// 交叉编译 Windows
// CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o excel-split-merge-v0.0.2.exe main.go

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	ToolVersion = "v0.0.2"
	ToolAuthor  = "AIslandX <yuchunyu97@gmail.com>"
)

// excel 第一行为标题
// 接下来的每一行会单独和标题行拆分到单独的 excel 中
// 需要指定拆分哪些列，以及拆分后的 excel 命名规则

type ExcelMergeInfo struct {
	name    string
	content []int
}

type NewComment struct {
	Author string `json:"author"`
	Text   string `json:"text"`
}

func exec() error {
	// 从控制台获取输入的需要被拆分的 excel
	var inputExcelFileName string
	fmt.Printf("请输入需要拆分的 Excel 文件名（该文件需要和 exe 在同一个目录下，且文件名中不能带空格）：")
	if _, err := fmt.Scanln(&inputExcelFileName); err != nil {
		return errors.New(fmt.Sprintf("输入错误 %s", err))
	}

	// 读取 Excel 文件
	f, err := excelize.OpenFile(inputExcelFileName)
	if err != nil {
		return errors.New(fmt.Sprintf("读取 Excel 文件 %s 失败 %s", inputExcelFileName, err))
	}

	sheetList := f.GetSheetList()
	sheetName := sheetList[0]
	log.Println("正在处理 Sheet", sheetName)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return errors.New(fmt.Sprintf("读取 Sheet %s 失败 %s", sheetName, err))
	}

	// 标题行
	headRow := rows[0]
	for idx, cell := range headRow {
		fmt.Printf("%d. %s\t", idx+1, cell)
		if (idx+1)%5 == 0 {
			fmt.Printf("\n")
		}
	}
	var inputTitleIndex string
	fmt.Printf("\n请选择拆分后文件名及被合并的列，输入标题栏序号：")
	if _, err := fmt.Scanln(&inputTitleIndex); err != nil {
		return errors.New(fmt.Sprintf("输入错误 %s", err))
	}
	titleIndex, err := strconv.Atoi(inputTitleIndex)
	if err != nil {
		return errors.New(fmt.Sprintf("%s 格式错误，请输入数字 %s", inputTitleIndex, err))
	}
	if titleIndex <= 0 || titleIndex > len(headRow) {
		return errors.New(fmt.Sprintf("序号不在范围内（%d - %d）", 1, len(headRow)))
	}

	// 创建文件夹
	resultDirName := fmt.Sprintf("result_%s", time.Now().Format("20060102150405"))
	if err = os.Mkdir(resultDirName, os.ModePerm); err != nil {
		return errors.New(fmt.Sprintf("创建结果输出文件夹 %s 出错 %s", resultDirName, err))
	}

	// 获取全部工作表中的批注
	comments := f.GetComments()
	commentsList := comments[sheetName]

	// 获取需要合并的行数据
	var mergeList []*ExcelMergeInfo
	for idx, item := range rows {
		if idx > 0 {
			key := item[titleIndex-1]
			isExist := false
			for _, mergeItem := range mergeList {
				if mergeItem.name == key {
					mergeItem.content = append(mergeItem.content, idx)
					isExist = true
					break
				}
			}
			if !isExist {
				mergeList = append(mergeList, &ExcelMergeInfo{
					name:    key,
					content: []int{idx},
				})
			}
		}
	}

	// 输出结果
	fmt.Printf("\n开始处理：\n\n")
	successCount, failedCount := 0, 0
	for idx, mergeInfo := range mergeList {
		log.Printf("开始处理第 %d 条数据\n", idx+1)

		resultFile := excelize.NewFile()
		newSheetName := "Sheet1"
		index := resultFile.NewSheet(newSheetName)
		resultFile.SetActiveSheet(index)

		// 复制样式表
		resultFile.Styles = f.Styles

		// 设置标题行行高
		row1Height, _ := f.GetRowHeight(sheetName, 1)
		_ = resultFile.SetRowHeight(newSheetName, 1, row1Height)
		// 设置标题行内容和列宽
		for i, val := range headRow {
			colName, _ := excelize.ColumnNumberToName(i + 1)

			// 设置列宽
			colWidth, _ := f.GetColWidth(sheetName, colName)
			_ = resultFile.SetColWidth(newSheetName, colName, colName, colWidth)

			// 设置内容
			_ = resultFile.SetCellValue(newSheetName, fmt.Sprintf("%s1", colName), val)
			// 设置样式
			styleHeader, _ := f.GetCellStyle(sheetName, fmt.Sprintf("%s1", colName))
			_ = resultFile.SetCellStyle(newSheetName, fmt.Sprintf("%s1", colName), fmt.Sprintf("%s1", colName), styleHeader)
		}

		for num, rowIdx := range mergeInfo.content {
			// 设置行高
			rowHeight, _ := f.GetRowHeight(sheetName, rowIdx+1)
			_ = resultFile.SetRowHeight(newSheetName, num+2, rowHeight)

			// 按行插入数据
			for i := range headRow {
				colName, _ := excelize.ColumnNumberToName(i + 1)
				row := rows[rowIdx]

				// 增加对 Excel 批注的复制功能
				originRef := fmt.Sprintf("%s%d", colName, rowIdx+1)
				var comment NewComment
				for _, c := range commentsList {
					if c.Ref == originRef {
						comment.Author = c.Author
						comment.Text = c.Text
					}
				}
				if comment != (NewComment{}) {
					// 插入批注
					b, err := json.Marshal(comment)
					if err != nil {
						log.Printf("\ncomment json.Marshal err: %s\n", err)
					}
					err = resultFile.AddComment(newSheetName, fmt.Sprintf("%s%d", colName, num+2), string(b))
					if err != nil {
						log.Printf("\ncomment resultFile.AddComment err: %s\n", err)
					}
				}

				// 设置内容
				value := ""
				if i < len(row) {
					value = row[i]
				}
				_ = resultFile.SetCellValue(newSheetName, fmt.Sprintf("%s%d", colName, num+2), value)
				// 设置样式
				styleContent, _ := f.GetCellStyle(sheetName, fmt.Sprintf("%s%d", colName, rowIdx+1))
				_ = resultFile.SetCellStyle(newSheetName, fmt.Sprintf("%s%d", colName, num+2), fmt.Sprintf("%s%d", colName, num+2), styleContent)
			}
		}

		// 生成文件名
		newExcelName := filepath.Join(resultDirName, fmt.Sprintf("%s.xlsx", mergeInfo.name))

		err := resultFile.SaveAs(newExcelName)
		if err != nil {
			log.Printf("失败：%s\n\n", err)
			failedCount++
		} else {
			log.Printf("成功：%s\n\n", newExcelName)
			successCount++
		}
	}

	log.Printf("处理完成，成功 %d 条，失败 %d 条\n\n", successCount, failedCount)

	return nil
}

func main() {
	fmt.Printf("欢迎使用 Excel 按行拆分小工具升级版\nversion %s\nauthor %s\n\n", ToolVersion, ToolAuthor)

	err := exec()
	if err != nil {
		log.Println("error:", err)
	}

	fmt.Printf("按任意键退出")
	_, _ = fmt.Scanln()
}
