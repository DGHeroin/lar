package larc

import (
    "archive/zip"
    "errors"
    "github.com/DGHeroin/golua/lua"
    "io"
    "io/ioutil"
    "log"
    "os"
    "path"
    "path/filepath"
    "strings"
    "sync"
)

var (
    ErrorLarFileNotFound      = errors.New("lar file not found")
    ErrorLuaFileNotFound      = errors.New("lua file not found")
    ErrorLarFileLoadDuplicate = errors.New("lar file load duplicate")
    ErrorSearchPathNotExists  = errors.New("search path not a dir")
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

// Lua Archive
type Lar struct {
    L           *lua.State
    mutexRun    sync.RWMutex
    mutexLoad   sync.RWMutex
    loaders     map[string]*loader
    searchPaths []string
}

var (
    PostNew func(*Lar)
)

// New Lar
func New(Ls ...*lua.State) *Lar {
    var L *lua.State
    if len(Ls) == 1 && Ls[0] != nil {
        L = Ls[0]
    }
    lar := &Lar{
        loaders: make(map[string]*loader),
    }
    if L == nil {
        L = lua.NewState()
    }
    L.OpenLibs()
    L.OpenGoLibs()
    L.PushGoStruct(lar)
    L.SetGlobal("LarContext")
    L.Register("lar_load", lload)

    _ = L.DoString(InitCode)
    lar.L = L
    if PostNew != nil {
        PostNew(lar)
    }
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
func (l *Lar) getBytes(luaFilename string) ([]byte, error) {
    for _, loader := range l.loaders {
        if data, err := loader.getBytes(luaFilename); err == nil && data != nil {
            return data, err
        }
    }
    return nil, ErrorLuaFileNotFound
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
        name4 := strings.TrimPrefix(name3, src)
        name5 := strings.TrimPrefix(name4, "/")

        fh.Name = name5

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
        log.Println("pack to", dst, path, "=>", fh.Name)
        return err
    })
}
func (l *Lar) LoadFiles(files []string) error {
    for _, file := range files {
        if err := l.loadFile(file); err != nil {
            return err
        }
    }
    return nil
}

// load lar form file disk
func (l *Lar) loadFile(file string) error {
    l.mutexLoad.Lock()
    defer l.mutexLoad.Unlock()

    if _, ok := l.loaders[file]; ok {
        return ErrorLuaFileNotFound
    }

    if _, err := os.Stat(file); err != nil {
        if os.IsNotExist(err) {
            return ErrorLarFileNotFound
        }
        return err
    }

    loader := newLarLoader()
    if err := loader.preloadLarFile(file); err != nil {
        return err
    }
    l.loaders[file] = loader
    return nil
}
func (l *Lar) loadFileSearch(filename string) ([]byte, error) {
    for _, dir := range l.searchPaths {
        p := path.Join(dir, filename)
        if FileExist(p) {
            return ioutil.ReadFile(p)
        }
    }

    return nil, ErrorLuaFileNotFound
}
func (l *Lar) LoadMemory(customName string, data []byte) error {
    l.mutexLoad.Lock()
    defer l.mutexLoad.Unlock()

    if _, ok := l.loaders[customName]; ok {
        return ErrorLuaFileNotFound
    }
    loader := newLarLoader()
    if err := loader.LoadMemory(data); err != nil {
        return err
    }
    l.loaders[customName] = loader
    return nil
}

// run lua from disk
func (l *Lar) RunDiskFile(filename string) error {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return err
    }
    return l.DoString(string(data))
}

// lua do file in lar
func (l *Lar) DoFile(filename string) error {
    codeBytes, err := l.getBytes(filename)
    if err != nil {
        codeBytes, err = l.loadFileSearch(filename)
    }
    if err != nil {
        return err
    }
    return l.DoString(string(codeBytes))
}

// lua do string
func (l *Lar) DoString(code string) error {
    l.mutexRun.RLock()
    defer l.mutexRun.RUnlock()
    return l.L.DoString(code)
}

func (l *Lar) AddSearchPaths(s ...string) error {
    for _, str := range s {
        if err := l.AddSearchPath(str); err != nil {
            return err
        }
    }
    return nil
}

// add lua search path in filesystem
func (l *Lar) AddSearchPath(s string) error {
    if s == "" {
        return ErrorSearchPathNotExists
    }

    if st, err := os.Stat(s); err != nil {
        if os.IsNotExist(err) {
            return nil
        }
        return err
    } else {
        if !st.IsDir() {
            return nil
        }
    }
    l.searchPaths = append(l.searchPaths, s)
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
    return nil
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
