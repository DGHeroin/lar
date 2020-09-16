package lar

import (
	"archive/zip"
	"bytes"
	"errors"
	"github.com/DGHeroin/golua/lua"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrorLarFileNotFound = errors.New("lar file not found")
	ErrorLuaFileNotFound = errors.New("lua file not found")
)

var (
	InitCode = `
local _lar_load = lar_load
lar_load = nil -- hide in global var
local function custom_loader( name )
    local code = _lar_load(name)
    if code then
        return load(code)
    end
end

table.insert(package.searchers, custom_loader)
-- disable load c lib
table.remove(package.searchers, 4)
table.remove(package.searchers, 3)

`
)
//Lua Archive
type Lar struct {
	z           *zip.ReadCloser
	readiedFile map[string][]byte
	fileReaders map[string]io.ReadCloser
	L           *lua.State
}
// New Lar
func New() *Lar {
	lar := &Lar{}
	lar.readiedFile = make(map[string][]byte)
	lar.fileReaders = make(map[string]io.ReadCloser)
	L := lua.NewState()
	L.OpenLibs()
	L.OpenGoLibs()
	L.PushGoStruct(lar)
	L.SetGlobal("LarContext")
	L.Register("lar_load", lload)

	_ = L.DoString(InitCode)
	lar.L = L
	return lar
}
// search lua file in lar
func lload(L *lua.State) int {
	name := L.ToString(1)
	L.GetGlobal("LarContext")
	_app := L.ToGoStruct(-1)
	if l, ok := _app.(*Lar); ok {
		luaFilename := strings.Replace(name, ".", "/", -1) + ".lua"
		if data, err := l.getBytes(luaFilename); err != nil || data == nil {
			return 0
		} else {
			l.L.PushString(string(data))
			return 1
		}
	}
	return 0
}
// pack dir all lua file to a lar file
func (l *Lar) Pack(dst, src string) error {
	fw, err := os.Create(dst)
	defer fw.Close()
	if err != nil {
		return err
	}
	zw := zip.NewWriter(fw)
	defer func() {
		_ = zw.Close()
	}()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fh, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		name1 := strings.TrimPrefix(path, src)
		name2 := strings.TrimPrefix(name1, string(filepath.Separator))
		name3 := strings.Replace(name2, "\\", "/", -1)
		fh.Name = name3

		if info.IsDir() {
			fh.Name += "/"
		}
		w, err := zw.CreateHeader(fh)
		if err != nil {
			return nil
		}
		if !fh.Mode().IsRegular() {
			return nil
		}
		if !strings.HasSuffix(path, ".lua") {
			return nil
		}
		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			return err
		}
		_, err = io.Copy(w, fr)
		log.Println("pack to", dst, path)
		return err
	})
}
// close lar
func (l *Lar) Close() {
	if l.z != nil {
		_ = l.z.Close()
	}
	l.z = nil
}
// load lar form file disk
func (l *Lar) LoadFile(file string, filename string) error {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return ErrorLarFileNotFound
		}
		return err
	}
	if err := l.preloadLarFile(file); err != nil {
		return err
	}
	return l.runFileInLar(filename)
}
// init lar from file
func (l *Lar) preloadLarFile(file string) error {
	zr, err := zip.OpenReader(file)
	if err != nil {
		return err
	}
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			continue
		}
		fr, err := file.Open()
		if err != nil {
			return err
		}
		l.fileReaders[file.Name] = fr
	}

	return nil
}
// load lar form memory
func (l *Lar) LoadMemory(data []byte, filename string) error {
	r := bytes.NewReader(data)
	l.preloadLarMemory(r, r.Size())
	return l.runFileInLar(filename)
}
// init memory lar
func (l *Lar) preloadLarMemory(r io.ReaderAt, sz int64) error {
	zr, err := zip.NewReader(r, sz)
	if err != nil {
		return err
	}
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			continue
		}
		fr, err := file.Open()
		if err != nil {
			return err
		}
		l.fileReaders[file.Name] = fr
	}

	return nil
}
// run lua from disk
func (l *Lar) RunDiskFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return l.runString(string(data))
}
// get bytes data in lar file
func (l *Lar) getBytes(filename string) ([]byte, error) {
	if cacheByte, ok := l.readiedFile[filename]; ok {
		return cacheByte, nil
	}
	if fr, ok := l.fileReaders[filename]; ok {
		if data, err := ioutil.ReadAll(fr); err != nil {
			return nil, err
		} else {
			l.readiedFile[filename] = data
			return data, nil
		}
	}
	return nil, ErrorLuaFileNotFound
}
// do file in lar
func (l *Lar) runFileInLar(filename string) error {
	data, err := l.getBytes(filename)
	if err != nil {
		return err
	}
	return l.runString(string(data))
}
// dostring
func (l *Lar) runString(code string) error {
	return l.L.DoString(code)
}
// add lua search path in filesystem
func (l *Lar) AddSearchPath(s string) {
	if s == "" {
		return
	}
	if !strings.HasSuffix(s, "/?.lua") {
		s = s + "/?.lua"
	}
	L := l.L
	L.GetGlobal("package")
	L.CheckType(1, lua.LUA_TTABLE)

	L.GetField(-1, "path")
	L.CheckType(-1, lua.LUA_TSTRING)
	luaPaths := L.CheckString(-1)

	// concat package.path + new_path
	paths := strings.Split(luaPaths, ";")
	paths = append(paths, s)
	paths = unique(paths)
	luaPaths = strings.Join(paths, ";")

	// set package.path
	L.GetGlobal("package")
	L.PushString("path")
	L.PushString(luaPaths)
	L.SetTable(-3)
	os.Setenv("LUA_PATH", "scripts/?.lua")
}
// make []string item unique
func unique(intSlice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
