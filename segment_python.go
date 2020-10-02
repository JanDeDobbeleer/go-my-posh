package main

import (
	"fmt"
	"strings"
)

type python struct {
	props         *properties
	env           environmentInfo
	venvName      string
	pythonVersion string
}

const (
	//DisplayVirtualEnv shows or hides the virtual env
	DisplayVirtualEnv Property = "display_virtual_env"
)

func (p *python) string() string {
	if p.venvName == "" || !p.props.getBool(DisplayVirtualEnv, true) {
		return p.pythonVersion
	}
	return fmt.Sprintf("%s %s", p.venvName, p.pythonVersion)
}

func (p *python) init(props *properties, env environmentInfo) {
	p.props = props
	p.env = env
}

func (p *python) enabled() bool {
	if !p.env.hasFiles("*.py") {
		return false
	}
	pythonVersions := []string{
		"python3",
		"python",
	}
	for index, python := range pythonVersions {
		version := p.env.runCommand(python, "--version")
		if version != "" {
			rawVersion := strings.TrimLeft(version, "Python")
			p.pythonVersion = strings.Trim(rawVersion, " ")
			break
		}
		//last element, Python isn't installed on this machine
		if index == len(pythonVersions)-1 {
			return false
		}
	}
	venvVars := []string{
		"VIRTUAL_ENV",
		"CONDA_ENV_PATH",
		"CONDA_DEFAULT_ENV",
		"PYENV_VERSION",
	}
	var venv string
	for _, venvVar := range venvVars {
		venv = p.env.getenv(venvVar)
		if venv != "" {
			p.venvName = base(venv, p.env)
			break
		}
	}
	return true
}
