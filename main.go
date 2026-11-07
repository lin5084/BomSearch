package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

var (
	workDir   string
	keyword   string
	foundList []string
	totalScan int
	startTime time.Time
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("用法：BomSearch 料号")
		fmt.Println("示例：./BomSearch SL-00001")
		os.Exit(1)
	}
	keyword = os.Args[1]
	startTime = time.Now()

	var err error
	workDir, err = os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("搜索目录：%s\n", workDir)
	fmt.Printf("搜索料号：%s\n", keyword)
	fmt.Println("----------------------------------------")

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
		if !strings.HasPrefix(strings.ToUpper(name), "SL") {
			continue
		}
		if !strings.HasSuffix(lowerName, ".xlsx") {
			continue
		}
		if strings.HasPrefix(name, "~$") {
			continue
		}

		totalScan++
		fullPath := filepath.Join(workDir, name)
		fmt.Printf("[处理中] %s\n", fullPath)

		hit := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("[PANIC 跳过] 文件异常：%s , err:%v\n", fullPath, r)
				}
			}()

			// 直接打开zip包，遍历内部xml文件查找文本
			z, err := excelize.OpenFile(fullPath)
			if err != nil {
				fmt.Printf("[跳过] 打开失败：%s  err:%v\n", fullPath, err)
				return
			}
			defer func() {
				_ = z.Close()
			}()

			// 遍历zip内所有文件，只读取xl下xml
			for _, e := range z.ZipReader.File {
				if !strings.HasPrefix(e.Name, "xl/") || !strings.HasSuffix(e.Name, ".xml") {
					continue
				}
				rc, err := e.Open()
				if err != nil {
					continue
				}
				buf, err := os.ReadFile(rc)
				_ = rc.Close()
				if err != nil {
					continue
				}
				if strings.Contains(string(buf), keyword) {
					hit = true
					break
				}
			}
		}()

		if hit {
			foundList = append(foundList, fullPath)
			fmt.Printf("====> [找到] %s\n", fullPath)
		}
	}

	duration := time.Since(startTime)

	// 控制台输出汇总
	fmt.Println("----------------------------------------")
	fmt.Printf("此次查询的料号为：%s，共查询了 %d 个文档，共有 %d 个文档包含此料号，查询耗时：%v\n",
		keyword, totalScan, len(foundList), duration)
	if len(foundList) == 0 {
		fmt.Println("未找到任何包含该料号的SL开头BOM文件")
	} else {
		fmt.Println("查到包含此物料的文档如下：")
		for _, s := range foundList {
			fmt.Println(s)
		}
	}

	// 写入结果txt
	outFile := filepath.Join(workDir, "BomSearch_Result.txt")
	f, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("无法写入结果文件：%v\n", err)
	} else {
		defer f.Close()
		w := bufio.NewWriter(f)
		line1 := fmt.Sprintf("此次查询的料号为%s，共查询了%d个文档，共有%d个文档包含此料号，此次查询耗时%v。查到包含此物料的文档如下。\n",
			keyword, totalScan, len(foundList), duration)
		_, _ = w.WriteString(line1)
		for _, s := range foundList {
			_, _ = w.WriteString(s + "\n")
		}
		_ = w.Flush()
		fmt.Printf("\n结果已保存到：%s\n", outFile)
	}
}
