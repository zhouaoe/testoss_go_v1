# testoss_go_v1

使用方法:
1 git clone https://github.com/zhouaoe/testoss_go_v1
2 编译： go build -o preformancetest
3 修改配置文件config.json
{
"endpoint": "xxxx", 域名
"bucketname": "xxx", 桶名
"write_progress": false, 是否上传测试数据
"read_progress": true, 是否执行下载测试
"read_range": 64, 下载测试是否是range读取，超过0表示range读X KB
"test_dir" : "1test0099/", 测试目录
"test_file_num": 10000, 测试文件个数
"threadNum": 8, 任务线程数据
"cleanData": false, 是否清理数据
"testFileSizeList": [64,1024, 4096] 测试文件大小，每个类型大小都有test_file_num个文件
}

4 设置环境变量
export OSS_ACCESS_KEY_ID=xxxxx
export OSS_ACCESS_KEY_SECRET=xxxxx

5 执行
./preformancetest  -config=config.json -debug=false
