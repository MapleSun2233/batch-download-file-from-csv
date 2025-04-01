package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func main() {
	csv_file := flag.String("csv", "", "csv file")
	charset := flag.String("charset", "UTF-8", "文本字符集，默认UTF-8")
	download_dir := flag.String("download_dir", "", "下载目录")
	start_line_index := flag.Int("start_line_index", 0, "开始行索引")
	file_name_index := flag.Int("file_name_index", 0, "文件名列索引")
	download_url_index := flag.Int("download_url_index", 0, "下载地址列索引")
	error_log := flag.String("error_log", "./failed.csv", "错误日志输出位置")
	flag.Parse()

	// 检查参数
	if *csv_file == "" {
		fmt.Println("请指定csv文件路径")
		return
	}
	if *download_dir == "" {
		fmt.Println("请指定下载目录")
		return
	}

	// 打印参数
	fmt.Printf(`params info：
	csv_file:%s
	charset:%s
	download_dir:%s
	start_line_index:%d
	file_name_index:%d
	download_url_index:%d
	error_log:%s
	`, *csv_file, *charset, *download_dir, *start_line_index, *file_name_index, *download_url_index, *error_log)
	// 检查以来文件夹是否存在
	_, err := os.Open(*download_dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("download_dir not exist")
		}
		return
	}
	// 准备依赖文件
	log, err := os.OpenFile(*error_log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	file, err := os.Open(*csv_file)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 读取csv文件
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()

	// 跳过无效标题行
	records = records[*start_line_index:]
	if err != nil {
		panic(err)
	}
	fmt.Println("reader handle files: ", len(records))

	// 准备字符转换器
	var decoder *encoding.Decoder
	switch strings.ToLower(*charset) {
	case "gbk":
		decoder = simplifiedchinese.GBK.NewDecoder()
	case "gb2312":
		decoder = simplifiedchinese.HZGB2312.NewDecoder()
	case "gb18030":
		decoder = simplifiedchinese.GB18030.NewDecoder()
	default:
	}

	// 初始化错误日志容器
	var failed_records [][]string = make([][]string, 1, len(records)+1)
	failed_records[0] = []string{"file_name", "download_url", "reason"}
	// 处理每一行
	for index, line := range records {
		file_name := line[*file_name_index]
		download_url := line[*download_url_index]
		if file_name == "" || download_url == "" {
			failed_records = append(failed_records, []string{file_name, download_url, "file_name or download_url is empty"})
			continue
		}
		if decoder != nil {
			file_name, err = decoder.String(file_name)
			if err != nil {
				failed_records = append(failed_records, []string{file_name, download_url, "fail to decode file_name"})
				continue
			}
		}
		fmt.Println("start download: ", file_name)
		// 准备文件下载
		download_file_path := *download_dir + "/" + file_name
		// 检查文件是否已存在
		if _, err := os.Stat(download_file_path); err == nil {
			// 如果文件已存在，添加时间戳
			timestamp := time.Now().UnixMilli()
			ext := path.Ext(file_name)
			base := strings.TrimSuffix(file_name, ext)
			file_name = fmt.Sprintf("%s_%d%s", base, timestamp, ext)
			download_file_path = *download_dir + "/" + file_name
		}
		download_file, err := os.OpenFile(download_file_path, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			failed_records = append(failed_records, []string{file_name, download_url, "fail to open file"})
			fmt.Println("fail to open file: ", download_file_path)
			continue
		}
		err = download(download_url, download_file)
		if err != nil {
			failed_records = append(failed_records, []string{file_name, download_url, "fail to download file"})
			fmt.Println("fail to download file: ", download_url)
			continue
		}
		download_file.Close()
		fmt.Println("download success: ", file_name, " handled: ", index+1, " of ", len(records))
	}
	// 写入错误日志
	writer := csv.NewWriter(log)
	writer.WriteAll(failed_records)
	writer.Flush()
}

func download(url string, file *os.File) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code is not 200")
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
