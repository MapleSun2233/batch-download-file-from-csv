# Filter Text Content
一个轻量级的文件下载批处理工具，数据来源为CSV，不使用Goroutine，不会对目标服务器造成访问压力。

# Build
```shell
go build -ldflags="-s -w" -o BatchDownloadFromCSV ./cmd/main.go
```
# Usage
```shell
Usage of ./BatchDownloadFromCSV:
  -csv string
        csv file
  -download_dir string
        下载目录
  -start_line_index int
        开始行索引
  -file_name_index int
        文件名列索引
  -download_url_index int
        下载地址列索引
  -error_log string
        错误日志输出位置 (default "./failed.csv")
  -charset string
        文本字符集，默认UTF-8 (default "UTF-8")
```
# Tips
1. start_line_index用于跳过可能存在的标题行。
2. charset用于解决中文乱码问题，可选值有`gbk`、`gb2312`、`gb18030`