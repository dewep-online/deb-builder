package control

import (
	"bytes"
	"os"
	"sort"

	"deb-builder/pkg/config"
	"deb-builder/pkg/utils"
)

type (
	Other struct {
		conf  *config.Config
		files []string
	}
	copyFile struct {
		Src string
		Dst string
	}
)

func NewOther(conf *config.Config) *Other {
	return &Other{
		conf:  conf,
		files: make([]string, 0),
	}
}

func (v *Other) WriteTo(dir string) error {
	buf := &bytes.Buffer{}

	if len(v.conf.Control.Conffiles) > 0 {
		conffilesFile := dir + "/conffiles"
		sort.Slice(v.conf.Control.Conffiles, func(i, j int) bool {
			return v.conf.Control.Conffiles[i] < v.conf.Control.Conffiles[j]
		})
		for _, filename := range v.conf.Control.Conffiles {
			if _, err := buf.WriteString(utils.FullPath(filename) + "\n"); err != nil {
				return err
			}
		}
		if err := os.WriteFile(conffilesFile, buf.Bytes(), 0644); err != nil {
			return err
		}
		buf.Reset()
		v.files = append(v.files, conffilesFile)
	}

	files := []copyFile{
		{Src: v.conf.Control.PreInstall, Dst: dir + "/preinst"},
		{Src: v.conf.Control.PostInstall, Dst: dir + "/postinst"},
		{Src: v.conf.Control.PreRemove, Dst: dir + "/prerm"},
		{Src: v.conf.Control.PostRemove, Dst: dir + "/postrm"},
	}

	for _, file := range files {
		if utils.FileExist(file.Src) {
			if err := utils.CopyFile(file.Src, file.Dst); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Other) List() []string {
	return v.files
}
