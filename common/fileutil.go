package common

import (
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
		return false, nil
	}
	return true, err
}

// 执行shell
func ExecShell(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)

	zap.S().Infof("run command: %s", cmd.String())

	out, err := cmd.CombinedOutput()

	return string(out), err
}

// 获取md5
func Md5V(str string) string {
	data := []byte(str)
	// nolint:gosec
	has := md5.Sum(data)
	md5str := fmt.Sprintf("%x", has)
	return md5str
}
