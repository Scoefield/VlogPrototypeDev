package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxByte = 10 * 1024 * 1024
	filePath = "./videos/"
)


//1. 输出hello
func sayHello(w http.ResponseWriter, r *http.Request)  {
	w.Write([]byte("hello world"))
}

func main() {
	//a.1 实现读取文件handler
	fileHandler := http.FileServer(http.Dir("./videos"))
	//a.2 注册handler
	http.Handle("/video/", http.StripPrefix("/video/", fileHandler))

	// 注册上传文件的handler
	http.HandleFunc("/api/upload", uploadHandler)

	// 注册获取视频文件列表接口的handler
	http.HandleFunc("/api/list", getFileListHandler)

	//2. 注册进server mux 就是将不同的url交给不同的handler
	http.HandleFunc("/hello", sayHello)
	//3. 启动监听端口，web服务
	err := http.ListenAndServe(":8090", nil)
	if err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}

//1. 业务逻辑，处理客户端上传的文件接口
func uploadHandler(w http.ResponseWriter, r *http.Request)  {
	//1.1 限制上传上来的视频文件大小: maxByte(10MB)
	r.Body = http.MaxBytesReader(w, r.Body, maxByte)
	//解析
	err := r.ParseMultipartForm(maxByte)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//1.2 获取上传的文件
	file, fileHeader, err := r.FormFile("uploadFile")

	//1.3 检查文件是否以mp4结尾
	ret := strings.HasSuffix(fileHeader.Filename, ".mp4")
	if ret == false {
		http.Error(w, "not mp4 file!", http.StatusInternalServerError)
		return
	}

	// 1.4 随机生成名称：文件名+时间戳
	md5Byte := md5.Sum([]byte(fileHeader.Filename + time.Now().String()))
	fmt.Println(md5Byte)
	// 对md5Byte进行16进制转换
	md5Str := fmt.Sprintf("%x", md5Byte)
	newFileName := md5Str + ".mp4"

	// 1.5 写入文件
	dst, err := os.Create(filePath + newFileName)
	defer dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

//2. 获取视频文件列表接口
func getFileListHandler(w http.ResponseWriter, r *http.Request)  {
	// 获取文件路径下的所有文件
	files, _ := filepath.Glob(filePath + "*")
	var ret []string
	for _, file := range files {
		fmt.Println(r.Host, file)
		ret = append(ret, "http://" + r.Host + "/video/"+ filepath.Base(file))
	}
	//返回json格式的url
	retJson, _ := json.Marshal(ret)
	w.Write(retJson)
	return
}

