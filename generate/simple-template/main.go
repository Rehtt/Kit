package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

var (
	fullPath = flag.String("path", "", "项目目录地址")
	mod      = flag.String("mod", "", "项目mod")
)

//go:embed template
var templateData embed.FS

func main() {
	flag.Parse()

	var tempData = struct {
		Mod string
	}{
		Mod: *mod,
	}

	entry, err := templateData.ReadDir("template")
	if err != nil {
		log.Fatalln(err)
	}

	var tmp bytes.Buffer

	rangeDir(entry, "template", func(path string) error {
		p := filepath.Join(*fullPath, path[len("template"):])
		fmt.Println(p)
		err = mkdir(p)
		if err != nil {
			log.Fatalln(p, "创建目录失败：", err)
		}
		return err
	}, func(path string, data []byte) error {
		p := filepath.Join(*fullPath, path[len("template"):])

		// 检测目录
		dir, _ := filepath.Split(p)
		err = mkdir(dir)
		if err != nil {
			log.Fatalln(dir, "创建目录失败：", err)
		}

		t, err := template.New("").Parse(string(data))
		if err != nil {
			log.Fatalln(p, "读取模板失败：", err)
		}

		tmp.Reset()
		err = t.ExecuteTemplate(&tmp, "", tempData)
		if err != nil {
			log.Fatalln(p, "模板拼装失败", err)
		}

		p = p[:len(p)-4]
		err = os.WriteFile(p, tmp.Bytes(), 0666)
		if err != nil {
			log.Fatalln(p, "写入文件失败：", err)
		}
		return nil
	})

	os.Chdir(*fullPath)
	c := exec.Command("go", "mod", "tidy")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	fmt.Println(c.Run())

	exec.Command("git", "add", ".").Run()

	log.Println("done")
}

// 遍历
func rangeDir(entry []fs.DirEntry, path string, onDir func(path string) error, onFile func(path string, data []byte) error) {
	for _, f := range entry {
		openPath := path + "/" + f.Name()
		if f.IsDir() {
			err := onDir(openPath)
			if err != nil {
				return
			}

			dir, err := templateData.ReadDir(openPath)
			if err != nil {
				log.Fatalln(openPath, "打开目录失败：", err)
			}
			rangeDir(dir, openPath, onDir, onFile) // 遍历目录
		} else {
			data, err := templateData.ReadFile(openPath)
			if err != nil {
				log.Fatalln(openPath, "读取失败：", err)
			}
			err = onFile(openPath, data)
			if err != nil {
				return
			}
		}
	}
}

func mkdir(p string) error {
	_, err := os.Stat(p)
	if err == nil {
		return nil
	}
	err = os.MkdirAll(p, 0777)
	return err
}
