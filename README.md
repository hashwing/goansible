# GoAnsible

使用golang 编写， 实现 ansible playbook 部分功能


## 编译

```

cd cmd/playbook/

go build -v

```

## 使用

```

playbook -workdir <your playbook dir> -tag <your tag>

```

## Host

```
host1 ansible_ssh_host=192.168.2.130 ansible_ssh_user=root ansible_ssh_pass=********** ansible_ssh_key=/root/.ssh/id_rsa
host2 ansible_ssh_host=192.168.2.132 ansible_ssh_user=root ansible_ssh_pass=********** ansible_ssh_key=/root/.ssh/id_rsa

[test]
host1 k8s_master=yes
host2

```

## Playbook

```yaml
- name: import playbook
  import_playbook: common.yaml

- name: test
  hosts: test
  vars:
    cluster_join_script: ""
    tests:
      - a: adddd
        b: b
      - a: c
        b: e
  include_values:
    - Values.yaml
  tasks:
    - name: debug value file
      shell: echo nnnn
      debug: "{{.Values.value_file}},{{ $test := index .Values.tests 0 }} {{ $test.a }}"
    - name: Copy file
      file:
        src: file.yaml
        dest: /root/file.yaml
    - name: test shell
      shell: echo {{ .HostVars.ansible_ssh_host }} > b.yml
    - name: stdout shell
      shell: echo {{ .HostVars.ansible_ssh_host }}
      stdout: hostvars.ip
    - name: stdout res
      shell: echo {{ .HostVars.ip }} > a.yaml
    - name: test when
      shell: echo hello > a.yaml
      when: hostvars.k8s_master
    - name: test template
      template:
        src: tpl.yaml
        dest: /root/tpl.yaml
    - name: get script
      shell: kubeadm token create --print-join-command
      stdout: values.cluster_join_script
      when: hostvars.k8s_master
    - name: stdout res
      shell: echo {{ .Values.cluster_join_script }} > join.yaml
      when: hostvars.k8s_master
    - name: regexp
      regexp:
        src: abrrrc
        exp: a(b{1})r(r{2})c
        dst: values.reg.ddd
      debug: "{{.Values.reg.ddd }}"
      
    - name: loop
      shell: echo "hello"
      loop:
        - a
        - b
    - name: loop 2
      shell: echo {{ .Item.a }} >> loop.yaml
      loop: values.tests
    - name: test ignore error
      shell: testddd
      ignore_error: true
    - name: setface
      setface: values.face=success
    - name: until 
      until:
        port: 3000
        timeout: 10
    - name: include tasks
      include: include.yaml

```