package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

var (
	workDir   string
	keyword   string
	foundList []string
)

func main() {
	// 参数：只接收1个参数（料号）
	if len(os.Args) != 2 {
		fmt.Println("用法：BomSearch 料号")
		fmt.Println("示例：./BomSearch SL-00001")
		os.Exit(1)
	}
	keyword = os.Args[1]

	// 获取程序所在目录（当前运行目录）
	var err error
	workDir, err = os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("搜索目录：%s\n", workDir)
	fmt.Printf("搜索料号：%s\n", keyword)
	fmt.Println("----------------------------------------")

	// 读取当前目录，不递归子文件夹
	entries, err := os.ReadDir(workDir)
	if err != nil {
		fmt.Printf("读取目录失败：%v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lowerName := strings.ToLower(name)
		// 条件：SL开头（忽略大小写） + .xlsx后缀 + 跳过临时~$文件
		if !strings.HasPrefix(strings.ToUpper(name), "SL") {
			continue
		}
		if !strings.HasSuffix(lowerName, ".xlsx") {
			continue
		}
		if strings.HasPrefix(name, "~$") {
			continue
		}

		fullPath := filepath.Join(workDir, name)
		fmt.Printf("[处理中] %s\n", fullPath)

		// 捕获panic，单个坏文件不卡死整个程序
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[PANIC 跳过] 文件异常：%s , err:%v\n", fullPath, r)
				}
			}()

			f, err := excelize.OpenFile(fullPath)
			if err != nil {
				fmt.Printf("[跳过] 打开失败：%s  err:%v\n", fullPath, err)
				return
			}
			defer func() {
				_ = f.Close()
			}()

			hit := false
			for _, sheet := range f.GetSheetList() {
				rows, err := f.Rows(sheet)
				if err != nil {
					fmt.Printf("[警告] 读取sheet失败：%s -> %s\n", fullPath, sheet)
					continue
				}
				for rows.Next() {
					row, err := rows.Columns()
					if err != nil {
						continue
					}
					for _, cellVal := range row {
						if strings.Contains(cellVal, keyword) {
							hit = true
							break
						}
					}
					if hit {
						break
					}
				}
				if hit {
					break
				}
			}

			if hit {
				foundList = append(foundList, fullPath)
				fmt.Printf("====> [找到] %s\n", fullPath)
			}
		}()
	}

	// 汇总输出
	fmt.Println("----------------------------------------")
	if len(foundList) == 0 {
		fmt.Println("未找到任何包含该料号的SL开头BOM文件")
	} else {
		fmt.Printf("一共找到 %d 个文件：\n", len(foundList))
		for _, s := range foundList {
			fmt.Println(s)
		}
	}

	// 结果写入当前目录 BomSearch_Result.txt
	outFile := filepath.Join(workDir, "BomSearch_Result.txt")
	f, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("无法写入结果文件：%v\n", err)
	} else {
		defer f.Close()
		w := bufio.NewWriter(f)
		for _, s := range foundList {
			_, _ = w.WriteString(s + "\n")
		}
		_ = w.Flush()
		fmt.Printf("\n结果已保存到：%s\n", outFile)
	}
}
