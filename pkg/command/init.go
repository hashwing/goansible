package command

import (
	"io/ioutil"
	"os"

	"github.com/hashwing/goansible/pkg/termutil"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const initValueContent = `
groups:
  all:
    localhost:
      ansible_ssh_host: ""
      ansible_ssh_port: ""
      ansible_ssh_user: ""
      ansible_ssh_pass: ""
      ansible_ssh_sudopass: ""
      ansible_ssh_key: ""
  test:
    localhost: {}
vars: 
  key1: value1

`
const initPlaybookContent = `
- name: example playbook
  hosts: all
  vars: {}
  tags:
  - d
  tasks:
    - name: echo
      shell: echo hello world
      stdout: values.name
      debug: '{{ .Values.name }}'
`

func newRunInitCmd() *cobra.Command {
	return &cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			err := ioutil.WriteFile("values.yaml", []byte(initValueContent), 0664)
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}
			err = ioutil.WriteFile("index.yaml", []byte(initPlaybookContent), 0664)
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}
			termutil.Successf("init finish, you can edit 'values.yaml' and 'index.yaml', then run ' ./goansible ' ")
		},
	}
}
