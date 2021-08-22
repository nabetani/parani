package main

import (
	_ "embed"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path"
	"sort"
	"strings"
	"time"
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
	if fn[0] == '.' {
		return true
	}
	switch strings.ToLower(path.Ext(fn)) {
	case ".png", ".jpeg", ".jpe", ".jpg":
		return false
	}
	return true
}

func sortedFileNames(files []fs.FileInfo) []string {
	fns := []string{}
	for _, f := range files {
		fns = append(fns, f.Name())
	}
	blocks := func(s string) []interface{} {
		r := []interface{}{}
		e := ""
		const (
			none  = iota
			num   = iota
			other = iota
		)
		t := none
		chtype := func(ch uint8) int {
			if '0' <= ch && ch <= '9' {
				return num
			}
			return other
		}
		add := func(e string) {
			switch t {
			case none:
				// do nothing
			case num:
				i := new(big.Int)
				i.SetString(e, 10)
				r = append(r, i)
			case other:
				r = append(r, e)
			}
		}
		for ix := 0; ix < len(s); ix++ {
			newT := chtype(s[ix])
			if newT == t {
				e += s[ix : ix+1]
			} else {
				add(e)
				e = s[ix : ix+1]
				t = newT
			}
		}
		add(e)
		return r
	}
	compare_obj := func(a, b interface{}) int {
		na, naOk := a.(*big.Int)
		nb, nbOk := b.(*big.Int)
		if naOk && nbOk {
			return na.Cmp(nb)
		}
		if naOk {
			return -1
		}
		if nbOk {
			return 1
		}
		sa, saOk := a.(string)
		sb, sbOk := b.(string)
		if saOk && sbOk {
			return strings.Compare(sa, sb)
		}
		log.Fatalf("a=%v b=%v", a, b)
		return 0
	}
	lessBlock := func(a, b []interface{}) bool {
		la := len(a)
		lb := len(b)
		l := func() int {
			if la < lb {
				return la
			}
			return lb
		}()
		for i := 0; i < l; i++ {
			c := compare_obj(a[0], b[i])
			if c != 0 {
				return c < 0
			}
		}
		return la < lb
	}

	sort.Slice(fns, func(i, j int) bool {
		bi := blocks(fns[i])
		bj := blocks(fns[j])
		return lessBlock(bi, bj)
	})
	return fns
}

func writeFile(s string) error {
	body := time.Now().Format("2006-01-02")
	for ix := 0; ; ix++ {
		fn := func() string {
			if ix == 0 {
				return body + ".html"
			}
			return fmt.Sprintf("%s_%d.html", body, ix)
		}()
		f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0664)
		if pe, ok := err.(*fs.PathError); pe != nil && ok && strings.Contains(pe.Err.Error(), "file exists") {
			continue
		}
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString(s); err != nil {
			return err
		}
		return nil
	}
}

func main() {
	imdir := "."
	files, err := ioutil.ReadDir(imdir)
	if err != nil {
		log.Fatal(err)
	}
	s := ""
	for _, fn := range sortedFileNames(files) {
		if isFileToIgnore(fn) {
			continue
		}
		url, err := base64img(path.Join(imdir, fn))
		if err != nil {
			log.Fatalln(err)
		}
		s += fmt.Sprintf("<img src='%s'/ name='%s'>", url, fn)
	}
	const key = "$$$image_tags$$$"
	if err := writeFile(strings.Replace(htmlString, key, s, 1)); err != nil {
		log.Fatal(err)
	}
}
