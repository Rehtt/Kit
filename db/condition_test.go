/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/8/6 下午 02:49
 */

package db

import (
	"fmt"
	"testing"
)

func TestCondition(t *testing.T) {
	type A struct {
		AA string `json:"a" gorm:"column:a" query:"=="`
	}
	var T struct {
		A
		ID   int    `json:"id" gorm:"column:id" query:"=="`
		Name string `json:"name" query:"like"`
		Test string `json:"test" query:"=="`
	}
	T.Name = "test"
	T.ID = 123
	T.AA = "test1"
	fmt.Println(Condition(&T))

	// out:
	// `a` = 'test1' AND `id` = '123' AND `Name` LIKE 'test'
}
