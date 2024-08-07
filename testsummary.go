package main

import "fmt"

type OssTestSummary struct {
	operationType    string
	operationName    string
	realTestDuration int64
	count            int
	requestDuration  []int64
}

// NewRequestTracker creates a new RequestTracker instance.
func NewOssTestSummary(operationType string, operationName string, count int) *OssTestSummary {
	return &OssTestSummary{
		operationType:    operationType,
		operationName:    operationName,
		count:            count,
		requestDuration:  make([]int64, count),
		realTestDuration: 0,
	}
}

func (summary *OssTestSummary) PrintSummary() {
	fmt.Printf("++++++ Run %s %s +++++ \n", summary.operationType, summary.operationName)
	fmt.Printf("OperationType: %s\n", summary.operationType)
	fmt.Printf("OperationName: %s\n", summary.operationName)
	fmt.Printf("Count: %d\n", summary.count)
	fmt.Printf("realTestDuration: %d\n", summary.realTestDuration)
	successCount := calculateSuccessCount(summary.requestDuration)
	fmt.Printf("success count: %d\n", successCount)
	fmt.Printf("err count: %d\n", summary.count-successCount)
	fmt.Printf("Average: %f\n", calculateAverage(summary.requestDuration))
}

func calculateSuccessCount(duration []int64) int {
	count := 0
	//排除为0的值，计算剩下的值的平均值
	for _, num := range duration {
		if num != 0 {
			count++
		}
	}
	return count

}

func calculateAverage(duration []int64) float64 {
	if len(duration) == 0 {
		return 0 // Avoid division by zero
	}

	var sum int64 = 0
	count := 0
	//排除为0的值，计算剩下的值的平均值
	for _, num := range duration {
		if num != 0 {
			sum += num
			count++
		}
	}

	return float64(sum) / float64(count)
}
