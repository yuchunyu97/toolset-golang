package main

// Excel Split

// 交叉编译 Windows
// CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o excel-split-v0.0.5.exe main.go

// Excelize 是 Go 语言编写的用于操作 Office Excel 文档基础库，基于 ECMA-376，ISO/IEC 29500 国际标准。
//     可以使用它来读取、写入由 Microsoft Excel™ 2007 及以上版本创建的电子表格文档。
//     支持 XLSX / XLSM / XLTM / XLTX 等多种文档格式，高度兼容带有样式、图片(表)、透视表、切片器等复杂组件的文档，并提供流式读写 API，
//     用于处理包含大规模数据的工作簿。可应用于各类报表平台、云计算、边缘计算等系统。使用本类库要求使用的 Go 语言为 1.15 或更高版本。
// docs: https://xuri.me/excelize/zh-hans/

import (
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	ToolVersion = "v0.0.5"
	ToolAuthor  = "AIslandX <yuchunyu97@gmail.com>"
)

// excel 第一行为标题
// 接下来的每一行会单独和标题行拆分到单独的 excel 中
// 需要指定拆分哪些列，以及拆分后的 excel 命名规则

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
	fmt.Printf("\n请选择拆分后文件名，输入标题栏序号（多个序号间用英文逗号分隔）：")
	if _, err := fmt.Scanln(&inputTitleIndex); err != nil {
		return errors.New(fmt.Sprintf("输入错误 %s", err))
	}

	inputTitleIndexList := strings.Split(inputTitleIndex, ",")
	var titleIndexList []int
	for _, item := range inputTitleIndexList {
		itemInt, err := strconv.Atoi(item)
		if err != nil {
			return errors.New(fmt.Sprintf("%s 格式错误，请输入数字 %s", item, err))
		}
		if itemInt <= 0 || itemInt > len(headRow) {
			return errors.New(fmt.Sprintf("序号 %d 不在范围内（%d - %d）", itemInt, 1, len(headRow)))
		}

		titleIndexList = append(titleIndexList, itemInt)
	}

	// 创建文件夹
	resultDirName := fmt.Sprintf("result_%s", time.Now().Format("20060102150405"))
	if err = os.Mkdir(resultDirName, os.ModePerm); err != nil {
		return errors.New(fmt.Sprintf("创建结果输出文件夹 %s 出错 %s", resultDirName, err))
	}

	// 输出结果
	fmt.Printf("\n开始处理：\n\n")
	successCount, failedCount := 0, 0
	for idx, row := range rows {
		if idx > 0 {
			log.Printf("开始处理第 %d 行数据\n", idx+1)

			resultFile := excelize.NewFile()
			newSheetName := "Sheet1"
			index := resultFile.NewSheet(newSheetName)
			resultFile.SetActiveSheet(index)

			// 复制样式表
			resultFile.Styles = f.Styles

			// 设置行高
			row1Height, _ := f.GetRowHeight(sheetName, 1)
			_ = resultFile.SetRowHeight(newSheetName, 1, row1Height)
			row2Height, _ := f.GetRowHeight(sheetName, idx+1)
			_ = resultFile.SetRowHeight(newSheetName, 2, row2Height)

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

				// 设置内容
				value := ""
				if i < len(row) {
					value = row[i]
				}
				_ = resultFile.SetCellValue(newSheetName, fmt.Sprintf("%s2", colName), value)
				// 设置样式
				styleContent, _ := f.GetCellStyle(sheetName, fmt.Sprintf("%s%d", colName, idx+1))
				_ = resultFile.SetCellStyle(newSheetName, fmt.Sprintf("%s2", colName), fmt.Sprintf("%s2", colName), styleContent)
			}

			// 生成文件名
			var nameList []string
			for _, nameIdx := range titleIndexList {
				nameList = append(nameList, row[nameIdx-1])
			}
			newExcelName := filepath.Join(resultDirName, fmt.Sprintf("%s.xlsx", strings.Join(nameList, "-")))

			err := resultFile.SaveAs(newExcelName)
			if err != nil {
				log.Printf("失败：%s\n\n", err)
				failedCount++
			} else {
				log.Printf("成功：%s\n\n", newExcelName)
				successCount++
			}
		}
	}

	log.Printf("处理完成，成功 %d 行，失败 %d 行\n\n", successCount, failedCount)

	return nil
}

func main() {
	fmt.Printf("欢迎使用 Excel 按行拆分小工具\nversion %s\nauthor %s\n\n", ToolVersion, ToolAuthor)

	err := exec()
	if err != nil {
		log.Println("error:", err)
	}

	fmt.Printf("按任意键退出")
	_, _ = fmt.Scanln()
}
