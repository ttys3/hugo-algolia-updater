package common

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"go.uber.org/zap"
)

// 读取文件
func ReadFileString(p string) string {
	b, _ := ioutil.ReadFile(p)
	return string(b)
}

// 写入文件
func WriteFile(path string, bytesArray []byte) {
	ioutil.WriteFile(path, bytesArray, 0o600)
}

// 判断是否存在
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		zap.S().Error("exists error: " + path + " not found")
		return false, nil
	}
	return true, err
}

// 执行shell
func ExecShell(name string, arg ...string) (string, error) {
	// 函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command(name, arg...)

	// 读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	err := cmd.Run()

	return out.String(), err
}

// 获取md5
func Md5V(str string) string {
	data := []byte(str)
	// nolint:gosec
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}
