package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
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

func pathToName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
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

func Check(err error) {
	if err != nil {
		log.Fatal("An error occurred: ", err)
	}
}

func printUsage(cli *Cli) func() {
	return func() {
		fmt.Fprint(os.Stdout, "Usage: engagor [OPTIONS] COMMAND\n\nA fancy script runner")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()

		help := "\nCommands:\n"

		for _, group := range cli.groups {
			help += fmt.Sprintf("	%-10.10s%s\n", group.name, group.description)
		}

		help += "\nRun 'engagor COMMAND --help' for more information on a command."
		fmt.Fprintf(os.Stdout, "%s\n", help)
	}
}

func main() {
	cli := NewCli()
	flag.Usage = printUsage(cli)
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	start := time.Now()
	err := cli.Cmd(flag.Args()...)
	Check(err)

	delta := time.Now().Sub(start)
	log.Printf("Took %0.3fs\n", delta.Seconds())
}
