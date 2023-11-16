/**
 * @Author: dsreshiram@gmail.com
 * @Date: 2022/8/6 下午 02:48
 */

package db

import (
	"fmt"
	"reflect"
	"strings"
)

// 简单转化查询条件
func Condition(src any) string {
	table := reflect.ValueOf(src)
	if table.Kind() == reflect.Ptr {
		table = table.Elem()
	}
	if table.Kind() != reflect.Struct {
		return ""
	}
	return condition(table)
}
func condition(value reflect.Value) string {
	var query []string
	ty := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := ty.Field(i)
		queryTag := fieldType.Tag.Get("query")
		if !field.IsZero() {
			if field.Kind() == reflect.Struct {
				query = append(query, condition(field))
				continue
			}
			if queryTag == "" {
				continue
			}
			// 获取字段名
			gormTag := fieldType.Tag.Get("gorm")
			if gormTag != "" {
				if index := strings.Index(gormTag, "column:"); index != -1 {
					gormTag = strings.Split(gormTag[index+len("column:"):], ";")[0]
				}
			} else {
				gormTag = fieldType.Name
			}

			switch queryTag {
			case "==":
				query = append(query, fmt.Sprintf("`%s` = '%v'", gormTag, field))
			case ">=":
				query = append(query, fmt.Sprintf("`%s` >= '%v'", gormTag, field))
			case "<=":
				query = append(query, fmt.Sprintf("`%s` <= '%v'", gormTag, field))
			case ">":
				query = append(query, fmt.Sprintf("`%s` > '%v'", gormTag, field))
			case "<":
				query = append(query, fmt.Sprintf("`%s` < '%v'", gormTag, field))
			case "!=":
				query = append(query, fmt.Sprintf("NOT `%s` = '%v'", gormTag, field))
			case "like":
				query = append(query, fmt.Sprintf("`%s` LIKE '%v'", gormTag, field))
			default:
				return ""
			}
		}
	}
	return strings.Join(query, " AND ")
}
