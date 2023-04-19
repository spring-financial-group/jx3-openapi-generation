package file

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spring-financial-group/mqa-logging/pkg/log"
	"io"
	"os"
	"path/filepath"
	"spring-financial-group/jx3-openapi-generation/pkg/domain"
	"spring-financial-group/jx3-openapi-generation/pkg/utils"
	"strings"
	"text/template"
)

type FileIO struct{}

func NewFileIO() domain.FileIO {
	return FileIO{}
}

func (f FileIO) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (f FileIO) ReplaceInFile(path, old, new string) error {
	bytes, err := f.Read(path)
	if err != nil {
		return errors.Wrap(err, "failed to read file")
	}

	data := strings.ReplaceAll(string(bytes), old, new)
	err = f.Write(path, []byte(data), 0700)
	if err != nil {
		return errors.Wrap(err, "failed to write file")
	}
	return nil
}

func (f FileIO) Copy(src string, dst string) (int64, error) {
	log.Logger().Info(fmt.Sprintf("%sCopying %s to %s%s", utils.Cyan, src, dst, utils.Reset))
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, errors.New(fmt.Sprintf("%s is not a regular file", src))
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	return io.Copy(destination, source)
}

func (f FileIO) CopyToDir(srcPath string, dstDir string) (int64, string, error) {
	dstPath := filepath.Join(dstDir, filepath.Base(srcPath))
	size, err := f.Copy(srcPath, dstPath)
	return size, dstPath, err
}

func (f FileIO) CopyManyToDir(dstDir string, srcFiles ...string) error {
	for _, file := range srcFiles {
		_, _, err := f.CopyToDir(file, dstDir)
		if err != nil {
			return errors.Wrapf(err, "failed to copy %s", filepath.Base(file))
		}
	}
	return nil
}

func (f FileIO) CopyToWorkingDir(srcPath string) (int64, error) {
	wd, err := os.Getwd()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get current working directory")
	}
	return f.Copy(srcPath, filepath.Join(wd, filepath.Base(srcPath)))
}

func (f FileIO) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (f FileIO) Write(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (f FileIO) Move(src string, dst string) error {
	return os.Rename(src, dst)
}

func (f FileIO) MkdirAll(path string, perm os.FileMode) (string, error) {
	err := os.MkdirAll(path, perm)
	return path, err
}

func (f FileIO) MkTmpDir(prefix string) (string, error) {
	return os.MkdirTemp("", prefix)
}

func (f FileIO) DeferRemove(path string) {
	if err := f.Remove(path); err != nil {
		log.Logger().Errorf("Failed to remove temporary directory: %s", err.Error())
	}
}

func (f FileIO) Remove(path string) error {
	return os.RemoveAll(path)
}

func (f FileIO) TemplateFiles(dstDir string, obj any, filePaths ...string) error {
	for _, path := range filePaths {
		if err := f.templateFile(dstDir, obj, path); err != nil {
			return err
		}
	}
	return nil
}

func (f FileIO) TemplateFilesInDir(srcDir, dstDir string, obj any) error {
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return errors.Wrap(err, "failed to read directory")
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if err = f.templateFile(dstDir, obj, filepath.Join(srcDir, file.Name())); err != nil {
			return err
		}
	}
	return nil
}

func (f FileIO) templateFile(dstDir string, obj any, filePath string) error {
	name := filepath.Base(filePath)
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		return errors.Wrapf(err, "failed to create template for %s", name)
	}

	path := filepath.Join(dstDir, filepath.Base(filePath))
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create new %s file", name)
	}
	defer file.Close()

	if err = tmpl.Execute(file, obj); err != nil {
		return errors.Wrapf(err, "failed to execute template for %s", name)
	}
	return nil
}
