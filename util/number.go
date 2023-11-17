package util

import "sort"

// FindMissingNumbers 寻找连续数字中的空缺数字
func FindMissingNumbers(numbers []int) []int {
	// 先对数字进行排序
	sort.Ints(numbers)

	var missingNumbers []int

	// 遍历数字切片，找出空缺数字
	for i := 1; i < len(numbers); i++ {
		if numbers[i] != numbers[i-1]+1 {
			// 发现空缺数字
			// missingRange := numbers[i-1] + 1
			for j := numbers[i-1] + 1; j < numbers[i]; j++ {
				missingNumbers = append(missingNumbers, j)
			}
			// fmt.Printf("空缺数字范围：%d 到 %d\n", missingRange, numbers[i]-1)
		}
	}
	return missingNumbers
}
