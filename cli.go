package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Cli struct {
	groups []Group
}

func NewCli() *Cli {
	cli := new(Cli)
	cli.initGroups()
	return cli
}

// read directories and build a list of command groups
// in order to qualify for a group, a directory:
//   - must have a README.md file
//   - must have scripts
func (cli *Cli) initGroups() error {
	top := "scripts"
	err := filepath.Walk(top, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() && path != top {
			group := NewGroup(path)
			cli.groups = append(cli.groups, *group)
			return filepath.SkipDir
		}
		return nil
	})
	return err
}

func (cli *Cli) FindGroup(name string) (*Group, error) {
	// find the command to execute
	for _, group := range cli.groups {
		if name == group.name {
			return &group, nil
		}
	}
	return nil, errors.New("Can't find group")
}

func (cli *Cli) showHelpForGroup(groupName string) {
	group, err := cli.FindGroup(groupName)
	if err != nil {
		return
	}
	help := group.name

	for _, script := range group.scripts {
		help += fmt.Sprintf("	%-20.20s%s\n", pathToName(script), script)
	}
	fmt.Printf("%s\n", help)
}

func (cli *Cli) Exec(groupName string, scriptName string) error {
	group, err := cli.FindGroup(groupName)
	if err != nil {
		return err
	}

	script, err := group.FindScript(scriptName)
	if err != nil {
		return err
	}

	log.Println("Executing", script)

	cmd := exec.Command(script)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	if err != nil {
		return errors.New(stderr.String())
	}

	fmt.Println(stdout.String())
	return nil
}

func (cli *Cli) Cmd(args ...string) error {
	switch count := len(args); count {
	case 1:
		cli.showHelpForGroup(args[0])
	case 2:
		// TODO: maybe splat args?
		cli.Exec(args[0], args[1])
	default:
		return errors.New("No group given!")
	}
	return nil
}
