package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

func handleError(err error) {
	myPrintf("Error:", err)
	os.Exit(-1)
}

var test_dir string = "test0099/"
var testFileSizeList []int
var testContentList = make(map[int][]byte)
var debug bool

func myPrintf(format string, a ...interface{}) {
	if debug {
		fmt.Printf(format, a...)
	}
}

//创建一个函数，输入为一个int，输出为test_prefix+int
func getTestFilepath(i int, sizeSuffix int) MyFileInfo {
	fileName := fmt.Sprintf("%s%08d_%d", test_dir, i, sizeSuffix)
	//myPrintf("gen fileName:", fileName)
	return NewMyFileInfo(fileName, i)
}

//func readFlag() *bool {
//	d := flag.Bool("debug", false, "debug flag")
//	flag.Parse()
//	fmt.Printf("flag d %t \n", *d)
//	return d
//
//}
func readJson() OSSTestConfig {
	// Define flag for the JSON configuration file path
	d := flag.Bool("debug", false, "debug flag")
	jsonFilePath := flag.String("config", "config.json", "Path to the JSON configuration file")

	flag.Parse()
	debug = *d

	// Read the JSON configuration file
	configBytes, err := ioutil.ReadFile(*jsonFilePath)
	if err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	// Unmarshal the JSON data into a Config struct
	var config OSSTestConfig
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %s", err)
	}

	// Print the configuration values
	myPrintf("Endpoint: %s\n", config.Endpoint)
	myPrintf("Bucket Name: %s\n", config.BucketName)
	myPrintf("Write Progress: %t\n", config.WriteProgress)
	myPrintf("Read Progress: %t\n", config.ReadProgress)
	myPrintf("Test File Number: %d\n", config.TestFileNum)
	myPrintf("Thread Number: %d\n", config.ThreadNum)
	myPrintf("Clean Data: %t\n", config.CleanData)
	myPrintf("Test File Size List: %v\n", config.TestFileSizeList)
	return config
}

func generate_test_data(testFileSizeList []int) {
	//循环testFileSizeList，以testFileSizeList的值为testContentList的key
	for _, size := range testFileSizeList {
		//生成一个size大小的数据到内存
		tempData := make([]byte, size)
		testContentList[size] = tempData
	}
	//打印testContentList的key
	//myPrintf("testContentList:", testContentList)
}
func main() {
	//debug = *readFlag()
	fmt.Printf("start test debug %t \n", debug)
	testConfig := readJson()
	//os.Exit(0)
	endpoint := testConfig.Endpoint
	bucketname := testConfig.BucketName
	write_progress := testConfig.WriteProgress
	read_progress := testConfig.ReadProgress
	test_file_num := testConfig.TestFileNum
	threadNum := testConfig.ThreadNum
	cleanData := testConfig.CleanData
	testFileSizeList = testConfig.TestFileSizeList
	generate_test_data(testFileSizeList)

	// 从环境变量中获取访问凭证。运行本代码示例之前，请确保已设置环境变量OSS_ACCESS_KEY_ID和OSS_ACCESS_KEY_SECRET。
	provider, err := oss.NewEnvironmentVariableCredentialsProvider()
	if err != nil {
		myPrintf("Error:", err)
		os.Exit(-1)
	}
	// 创建OSSClient实例。
	// yourEndpoint填写Bucket对应的Endpoint，以华东1（杭州）为例，填写为https://oss-cn-hangzhou.aliyuncs.com。其它Region请按实际情况填写。
	client, err := oss.New(endpoint, "", "", oss.SetCredentialsProvider(&provider))
	if err != nil {
		myPrintf("Error:", err)
		os.Exit(-1)
	}
	myPrintf("client:%#v\n", client)

	// 填写存储空间名称，例如examplebucket。
	bucket, err := client.Bucket(bucketname)
	if err != nil {
		myPrintf("Error:", err)
		os.Exit(-1)
	}

	if write_progress == true {
		for key, value := range testContentList {
			summary := NewOssTestSummary("upload", fmt.Sprintf("Object Size %d", key), test_file_num)
			myPrintf("---------write data : filesize=%d threadNum=%d test_file_num=%d \n", key, threadNum, test_file_num)
			writeData(bucket, threadNum, test_file_num, value, key, summary)
			summary.PrintSummary()
		}
	}

	if read_progress == true {
		for key, _ := range testContentList {
			summary := NewOssTestSummary("download", fmt.Sprintf("Object Size %d", key), test_file_num)
			myPrintf("---------read data : filesize=%d threadNum=%d test_file_num=%d \n", key, threadNum, test_file_num)
			readData(bucket, threadNum, test_file_num, key, summary)
			summary.PrintSummary()

		}
	}

	if cleanData == true {
		myPrintf("---------clean data ")
		cleanAllData(bucket)
	}

	//// 依次填写Object的完整路径（例如exampledir/exampleobject.txt）和本地文件的完整路径（例如D:\\localpath\\examplefile.txt）。
	//err = bucket.PutObjectFromFile("exampledir/exampleobject.txt", "D:\\localpath\\examplefile.txt")
	//if err != nil {
	//	myPrintf("Error:", err)
	//	os.Exit(-1)
	//}

}

func readData(bucket *oss.Bucket, threadNum int, testFileNum int, size int, summary *OssTestSummary) {
	startTime := time.Now()

	//用一个线程池，并发度为threadNum，上传test_file_num个文件
	bufferSize := 1024
	fileCh := make(chan MyFileInfo, bufferSize)
	var wg sync.WaitGroup

	// Start the worker goroutines.
	for i := 0; i < threadNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for fileInfo := range fileCh {
				DownloadFile(bucket, fileInfo, &summary.requestDuration[fileInfo.Index])
			}
		}()
	}

	// Send file names to the channel.
	go func() {
		for i := 0; i < testFileNum; i++ {
			fileCh <- getTestFilepath(i, size)
		}
		close(fileCh)
	}()

	wg.Wait()
	elapsedTime := time.Since(startTime)
	myPrintf("All files have been download. elapsedTime %d ms.\n", elapsedTime.Milliseconds())
}

func cleanAllData(bucket *oss.Bucket) {
	// List objects in the directory.
	startTime := time.Now()

	marker := ""
	isTruncated := true
	deletedNum := 0
	for isTruncated {
		objectList, err := bucket.ListObjects(oss.Prefix(test_dir), oss.Marker(marker))
		if err != nil {
			myPrintf("Failed to list objects: %s", err)
			return
		}

		// Process each object.
		for _, object := range objectList.Objects {
			//myPrintf("Deleting object: %s\n", object.Key)
			deleteObject(bucket, object.Key)
			deletedNum = deletedNum + 1
		}

		marker = objectList.NextMarker
		isTruncated = objectList.IsTruncated
	}
	elapsedTime := time.Since(startTime)

	myPrintf("All objects in the directory have been deleted. deletedNum=%d ,elapsedTime(ms)=%d \n", deletedNum, elapsedTime.Milliseconds())
}

func deleteObject(bucket *oss.Bucket, objectKey string) error {
	startTime := time.Now()

	err := bucket.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", objectKey, err)
	} else {
		elapsedTime := time.Since(startTime)
		myPrintf("#delete #FieName %s #elapsedTime(ms) %d \n", objectKey, elapsedTime.Milliseconds())
	}
	return nil
}

// getGoroutineID returns the current goroutine ID.
func getGoroutineID() (prefix string, id int64, suffix string) {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := fmt.Sprintf("%s", buf[:n])
	prefix, id, surfix := parseGoroutineID(idField)
	return prefix, id, surfix
}

// parseGoroutineID parses the goroutine ID from the stack trace.
func parseGoroutineID(stack string) (prefix string, id int64, suffix string) {
	// Find the goroutine ID in the stack trace.
	start := strings.Index(stack, "goroutine ")
	if start == -1 {
		return "", 0, ""
	}

	// Extract the goroutine ID.
	end := strings.Index(stack[start:], " ")
	if end == -1 {
		return "", 0, ""
	}

	// Parse the goroutine ID.
	idStr := stack[start+len("goroutine ") : start+len("goroutine ")+end]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return "", 0, ""
	}

	// Return the parsed goroutine ID.
	return stack[:start], id, stack[start+len("goroutine "+idStr)+end:]
}

func DownloadFile(bucket *oss.Bucket, fileInfo MyFileInfo, i *int64) {
	// 将内存中的数据tmpData上传到OSS
	fileName := fileInfo.FileName
	//reader := bytes.NewReader(tmpData)
	startTime := time.Now()
	content, err := bucket.GetObject(fileName)
	if err != nil {
		myPrintf("Error: File upload ", fileName, err)
	} else {
		//读取content的内容到内存
		//defer content.Close()
		// Discard the object data.
		_, err = io.Copy(ioutil.Discard, content)
		if err != nil {
			fmt.Errorf("failed to discard object data: %w", err)
		}

		elapsedTime := time.Since(startTime)
		*i = elapsedTime.Milliseconds()
		myPrintf("#download #FieName %s #elapsedTime(ms) %d \n", fileName, elapsedTime.Milliseconds())
	}
}

func uploadFile(bucket *oss.Bucket, fileInfo MyFileInfo, tmpData []byte, i *int64) error {
	fileName := fileInfo.FileName

	//prefix, gid, surfix := getGoroutineID()
	//myPrintf("Goroutine ID: %d ", gid)
	//将内存中的数据tmpData上传到OSS
	reader := bytes.NewReader(tmpData)
	startTime := time.Now()
	err := bucket.PutObject(fileName, reader)
	if err != nil {
		myPrintf("Error: File upload ", fileName, err)
	} else {
		elapsedTime := time.Since(startTime)
		*i = elapsedTime.Milliseconds()
		myPrintf("#upload #FieName %s #elapsedTime(ms) %d  \n", fileName, elapsedTime.Milliseconds())
	}
	return nil
}

func writeData(bucket *oss.Bucket, threadNum int, testFileNum int, tmpData []byte, size int, summary *OssTestSummary) {
	startTime := time.Now()

	//用一个线程池，并发度为threadNum，上传test_file_num个文件
	bufferSize := 1024
	fileCh := make(chan MyFileInfo, bufferSize)
	var wg sync.WaitGroup

	// Start the worker goroutines.
	for i := 0; i < threadNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for fileInfo := range fileCh {
				uploadFile(bucket, fileInfo, tmpData, &summary.requestDuration[fileInfo.Index])
			}
		}()
	}

	// Send file names to the channel.
	go func() {
		for i := 0; i < testFileNum; i++ {
			fileCh <- getTestFilepath(i, size)
		}
		close(fileCh)
	}()

	wg.Wait()
	elapsedTime := time.Since(startTime)
	myPrintf("All files have been uploaded. elapsedTime %d ms.\n", elapsedTime.Milliseconds())
	summary.realTestDuration = elapsedTime.Milliseconds()
}
