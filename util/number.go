package util

import (
	"sort"
)

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

// GroupNumbers 将连续数字进行分组
// numbers 输入数字
// groupSize 每组最多容纳数字个数，当groupSize <= 0时不做限制
func GroupNumbers(numbers []int, groupSize int) [][]int {
	var result [][]int
	var currentGroup []int

	for i := 0; i < len(numbers); i++ {
		if i > 0 && numbers[i] != numbers[i-1]+1 && len(currentGroup) > 0 {
			result = append(result, currentGroup)
			currentGroup = []int{}
		}
		currentGroup = append(currentGroup, numbers[i])

		if groupSize > 0 && len(currentGroup) == groupSize && len(currentGroup) > 0 {
			result = append(result, currentGroup)
			currentGroup = []int{}
		}
	}
	if len(currentGroup) > 0 {
		result = append(result, currentGroup)
	}
	return result
}
