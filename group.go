package main

import (
	"bufio"
	"errors"
	"os"
	"path"
	"path/filepath"
)

type Group struct {
	name        string
	path        string
	description string
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

func (group *Group) initDescription() error {
	readme := path.Join(group.path, "README.md")
	file, err := os.Open(readme)
	if err != nil {
		return err
	}
	defer file.Close()

	r := bufio.NewReader(file)
	line, err := r.ReadString('\n')
	Check(err)

	group.description = line[2 : len(line)-1]
	return nil
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
