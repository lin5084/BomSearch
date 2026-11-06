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
	searchRoot string
	keyword    string
	foundList  []string
)

func main() {
	// 参数校验
	if len(os.Args) != 3 {
		fmt.Println("用法：BomSearch.exe <BOM文件夹路径> <要查找的料号>")
		fmt.Println("示例：BomSearch.exe D:\\BOM_FOLDER ITEM-001234")
		os.Exit(1)
	}

	searchRoot = os.Args[1]
	keyword = os.Args[2]

	// 判断目录是否存在
	stat, err := os.Stat(searchRoot)
	if err != nil {
		fmt.Printf("错误：目录不存在 %s \n", searchRoot)
		os.Exit(1)
	}
	if !stat.IsDir() {
		fmt.Printf("错误：%s 不是文件夹\n", searchRoot)
		os.Exit(1)
	}

	fmt.Printf("开始搜索，根目录：%s，料号：%s\n", searchRoot, keyword)
	fmt.Println("----------------------------------------")

	// 遍历目录
	err = filepath.Walk(searchRoot, walkFile)
	if err != nil {
		fmt.Printf("遍历目录异常: %v\n", err)
	}

	// 输出结果到控制台
	fmt.Println("----------------------------------------")
	if len(foundList) == 0 {
		fmt.Println("未找到任何包含该料号的BOM文件")
	} else {
		fmt.Printf("一共找到 %d 个文件：\n", len(foundList))
		for _, s := range foundList {
			fmt.Println(s)
		}
	}

	// 写入结果文件
	outFile := filepath.Join(searchRoot, "BomSearch_Result.txt")
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

func walkFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	// 只处理xlsx，跳过临时文件 ~$开头
	name := strings.ToLower(info.Name())
	if !strings.HasSuffix(name, ".xlsx") {
		return nil
	}
	if strings.HasPrefix(info.Name(), "~$") {
		return nil
	}

	// 打开excel，只读模式
	f, err := excelize.OpenFile(path, excelize.Options{ReadOnly: true})
	if err != nil {
		fmt.Printf("[跳过] 读取失败：%s  err:%v\n", path, err)
		return nil
	}
	defer func() {
		_ = f.Close()
	}()

	// 遍历所有sheet
	hit := false
	for _, sheet := range f.GetSheetList() {
		rows, err := f.Rows(sheet)
		if err != nil {
			continue
		}
		for rows.Next() {
			row, err := rows.Columns()
			if err != nil {
				continue
			}
			for _, cellVal := range row {
				if cellVal == keyword {
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
		foundList = append(foundList, path)
		fmt.Printf("[找到] %s\n", path)
	}
	return nil
}
