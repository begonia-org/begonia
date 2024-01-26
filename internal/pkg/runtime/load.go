package runtime

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wetrycode/begonia/internal/pkg/config"
)

type ProtosLoader interface {
	LoadProto(zipFile string, tmpDir string, pkg string, out string) error
}

type protoLoaderImpl struct {
	config *config.Config
	lock   sync.RWMutex
}

func NewProtoLoaderImpl(config *config.Config) ProtosLoader {
	return &protoLoaderImpl{
		config: config,
		lock:   sync.RWMutex{},
	}
}
func (p *protoLoaderImpl) unzip(zipFile string, destDir string) error {
	// 打开 zip 文件
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()

	// 遍历压缩文件中的每个文件/目录
	for _, f := range r.File {
		// 计算目标路径
		fpath := filepath.Join(destDir, f.Name)

		// 检查文件是否为目录
		if f.FileInfo().IsDir() {
			// 创建目录
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// 创建文件的所有上级目录
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// 打开压缩包内的文件
		inFile, err := f.Open()
		if err != nil {
			return err
		}
		defer inFile.Close()

		// 创建要写入的文件
		outFile, err := os.Create(fpath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		// 将文件内容复制到新文件
		if _, err = io.Copy(outFile, inFile); err != nil {
			return err
		}
	}
	return nil
}

// CopyFile 复制文件的函数
func (p *protoLoaderImpl) CopyFile(src, dst string) error {
	// 打开源文件
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 使用 io.Copy 复制内容
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 确保所有内容都写入目标文件
	err = destFile.Sync()
	return err
}
func (p *protoLoaderImpl) Makefile(dir, pkg, outDir string) error {
	cmd := exec.Command("make", "go", "PKG="+pkg, "OUT="+outDir)
	// 创建一个缓冲区用来存储命令的输出
	var out bytes.Buffer
	cmd.Stdout = &out
	wd, _ := os.Getwd()
	log.Println("执行命令:", cmd.String(), wd)
	// 执行命令
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return fmt.Errorf("执行命令失败,%w,%s", err, string(out.Bytes()))
	}
	return nil
}
func (p *protoLoaderImpl) LoadProto(zipFile string, tmpDir string, pkg string, out string) error {
	filename := filepath.Base(zipFile)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	dir := filepath.Join(tmpDir, filename, time.Now().Format("20060102150405"), "protos")

	err := p.unzip(zipFile, dir)
	if err != nil {
		return fmt.Errorf("解压文件失败,%w", err)
	}
	err = p.CopyFile("/data/work/wetrycode/begonia/internal/pkg/runtime/Makefile", filepath.Join(dir, "Makefile"))
	if err != nil {
		return err
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	err = os.Chdir(dir)
	if err != nil {
		return fmt.Errorf("切换目录失败,%w", err)
	}
	return p.Makefile(dir, pkg, out)
}
