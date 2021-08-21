package main

import (
	_ "embed"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

//go:embed html/index.html
var htmlString string

func mimeText(fn string) (string, error) {
	switch strings.ToLower(path.Ext(fn)) {
	case ".png":
		return "image/png", nil
	case ".jpg", ".jpeg", ".jpe":
		return "image/jpeg", nil
	}
	return "", errors.New("unexpected file type: " + fn)
}

func base64img(fn string) (string, error) {
	bytes, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}
	enc := b64.StdEncoding.EncodeToString(bytes)
	mime, err := mimeText(fn)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("data:%s;base64,%s", mime, enc), nil
}
func isFileToIgnore(fn string) bool {
	log.Println(fn, " ", fn[0])
	if fn[0] == '.' {
		return true
	}
	switch strings.ToLower(fn) {
	case "thumbs.db", "ehthumbs.db", "ehthumbs_vista.db", "desktop.ini", "icon":
		return true
	}
	switch strings.ToLower(path.Ext(fn)) {
	case ".stackdump", ".lnk":
	}
	return false
}

func main() {
	log.Println(htmlString)
	dir, _ := path.Split(os.Args[0])
	imdir := path.Join(dir, "html/sample_images")
	files, err := ioutil.ReadDir(imdir)
	if err != nil {
		log.Fatal(err)
	}
	s := ""
	for _, file := range files {
		if isFileToIgnore(file.Name()) {
			continue
		}
		url, err := base64img(path.Join(imdir, file.Name()))
		if err != nil {
			log.Fatalln(err)
		}
		s += fmt.Sprintf("<img src='%s'/>", url)
	}
	const key = "$$$image_tags$$$"
	replaced := strings.Replace(htmlString, key, s, 1)
	fmt.Println(replaced)
}
