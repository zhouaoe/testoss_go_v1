package main

type OSSTestConfig struct {
	Endpoint         string `json:"endpoint"`
	BucketName       string `json:"bucketname"`
	WriteProgress    bool   `json:"write_progress"`
	ReadProgress     bool   `json:"read_progress"`
	TestFileNum      int    `json:"test_file_num"`
	ThreadNum        int    `json:"threadNum"`
	CleanData        bool   `json:"cleanData"`
	TestFileSizeList []int  `json:"testFileSizeList"`
	ReadRange        int    `json:"read_range"`
	TestDir          string `json:"test_dir"`
}

type MyFileInfo struct {
	FileName string
	Index    int
	FileSize int
}

func NewMyFileInfo(fileName string, index int, fileSize int) MyFileInfo {
	return MyFileInfo{FileName: fileName, Index: index, FileSize: fileSize}
}
