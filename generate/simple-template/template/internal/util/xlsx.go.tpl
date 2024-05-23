package uilt

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"time"

	"github.com/tealeg/xlsx"
	"github.com/xuri/excelize/v2"
)

const (
	XlsxMax = 500000
)

type Xlsx struct {
	writer    *excelize.StreamWriter
	file      *excelize.File
	title     []interface{}
	index     int
	path      string
	fileIndex int
	fileName  string
}

func NewXlsx(filePath string, title interface{}) (x *Xlsx, err error) {
	path, fileName := filepath.Split(filePath)
	x = &Xlsx{
		path:     path,
		fileName: fileName,
	}
	if err = x.parseTitle(title); err != nil {
		return nil, err
	}
	return
}

func (x *Xlsx) parseTitle(title interface{}) error {
	dataType := reflect.TypeOf(title)
	dataList := reflect.ValueOf(title)

	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
		dataList = dataList.Elem()
	}
	rangStruct := func() {
		for i := 0; i < dataType.NumField(); i++ {
			x.title = append(x.title, dataType.Field(i).Tag.Get("xlsx"))
		}
	}
	err := errors.New("目前仅支持 []string []struct struct")

	switch dataType.Kind() {
	case reflect.Array:
		for i := 0; i < dataList.Len(); i++ {
			switch dataList.Index(i).Kind() {
			case reflect.String: // []string
				x.title = append(x.title, dataList.Index(i).String())
			case reflect.Struct: // []struct
				rangStruct()
				return nil
			default:
				return err
			}
		}
	case reflect.Struct: // struct
		rangStruct()
	default:
		return err
	}
	return nil
}

func (x *Xlsx) Add(data interface{}) {
	if x.index == 0 {
		x.newFile()
		x.writer.SetRow("A1", x.title)
	}
	x.index += 1
	cell, _ := excelize.CoordinatesToCellName(1, x.index+1)
	x.writer.SetRow(cell, x.toArray(data))

	if x.index == XlsxMax {
		x.save()
	}
}

func (x *Xlsx) Close() {
	if x.index != 0 {
		x.save()
	}
}

func (x *Xlsx) save() {
	x.writer.Flush()
	x.file.SaveAs(x.getFileFullPath())
	x.index = 0
}

func (x *Xlsx) newFile() {
	x.file = excelize.NewFile()
	x.writer, _ = x.file.NewStreamWriter("Sheet1")
}

func (x *Xlsx) getFileFullPath() string {
	var tmp string
	if x.index == XlsxMax {
		x.fileIndex += 1
		ext := filepath.Ext(x.fileName)
		tmp = fmt.Sprintf("%s-%d%s", x.fileName[:len(x.fileName)-len(ext)], x.fileIndex, ext)
	} else {
		tmp = x.fileName
	}
	return filepath.Join(x.path, tmp)
}

func (x *Xlsx) toArray(data interface{}) (out []interface{}) {
	dataList := reflect.ValueOf(data)
	dataType := reflect.TypeOf(data)
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
		dataList = dataList.Elem()
	}
	switch dataList.Kind() {
	case reflect.Array:
		out = make([]interface{}, dataList.Len())
		for j := 0; j < dataList.Len(); j++ {
			out[j] = x.format(dataList.Index(j), "")
		}
	case reflect.Struct:
		out = make([]interface{}, dataList.NumField())
		for j := 0; j < dataList.NumField(); j++ {
			out[j] = x.format(dataList.Field(j), dataType.Field(j).Tag.Get("format"))
		}
	}
	return
}

func (x *Xlsx) format(v reflect.Value, foramt string) (out interface{}) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	switch v := v.Interface().(type) {
	case time.Time:
		if v.IsZero() {
			return ""
		}
		if foramt != "" {
			return v.Format(foramt)
		}
		return v
	default:
		return v
	}
}

// SaveFile 写入并导出文件
func SaveFile(data interface{}, fileName string) {
	dataType := reflect.TypeOf(data)
	dataList := reflect.ValueOf(data)
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
		dataList = dataList.Elem()
	}
	if dataType.Kind() != reflect.Slice {
		return
	}
	dataElemType := dataType.Elem()
	if dataElemType.Kind() == reflect.Ptr {
		dataElemType = dataElemType.Elem()
	}
	if dataElemType.Kind() != reflect.Struct {
		return
	}

	file := xlsx.NewFile()

	sheet, err := file.AddSheet("data")
	if err != nil {
		fmt.Printf("add sheet name failed,err=%s\n", err.Error())
		return
	}

	title := sheet.AddRow()
	for i := 0; i < dataElemType.NumField(); i++ {
		titleTag := dataElemType.Field(i).Tag.Get("xlsx")
		if titleTag == "" {
			titleTag = dataElemType.Field(i).Name
		}
		title.AddCell().SetValue(titleTag)
	}

	for i := 0; i < dataList.Len(); i++ {
		row := sheet.AddRow()
		dataElemList := dataList.Index(i)
		if dataElemList.Kind() == reflect.Ptr {
			dataElemList = dataElemList.Elem()
		}
		for j := 0; j < dataElemList.NumField(); j++ {
			dataElem := dataElemList.Field(j)
			if dataElem.Kind() == reflect.Ptr {
				dataElem = dataElem.Elem()
			}
			if dataElem.Kind() == reflect.Invalid {
				row.AddCell().SetString("")
				continue
			}
			switch data := dataElem.Interface().(type) {
			case time.Time:
				if form := dataElemType.Field(j).Tag.Get("form"); form != "" {
					row.AddCell().SetString(data.Format(form))
				}
			case int:
				row.AddCell().SetString(fmt.Sprintf("%d", data))
			default:
				row.AddCell().SetValue(data)
			}
		}
	}
	err = file.Save(fileName)
	if err != nil {
		fmt.Printf("save file failed,err:%s\n", err.Error())
		return
	}
	fmt.Println("export " + fileName + " success\n")
}
