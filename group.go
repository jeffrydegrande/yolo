package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Group struct {
	name        string
	path        string
	description string
	help        string
	scripts     []string
}

func NewGroup(path string) *Group {
	group := new(Group)
	group.name = filepath.Base(path)
	group.path = path
	group.initDescription()
	group.initScripts()
	return group
}

func (group *Group) initDescription() {
	path := path.Join(group.path, "README.md")
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	readme := string(file)
	pos := strings.Index(readme, "\n")
	group.description = readme[1 : pos-1]
	group.help = readme
}

func (group *Group) initScripts() error {
	return filepath.Walk(group.path, func(path string, f os.FileInfo, err error) error {
		mode := f.Mode()
		// only regular, executable files
		if mode.IsRegular() && (mode&0111 != 0) {
			group.scripts = append(group.scripts, path)
		}
		return nil
	})
}

func (group *Group) FindScript(scriptName string) (string, error) {
	for _, script := range group.scripts {
		if scriptName == pathToName(script) {
			return script, nil
		}
	}
	return "", errors.New("Can't find script")
}
