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

## 使用教程

goansible playbook格式跟 ansible playbook非常相似，但 goansible 没有role 功能，模板由 jinja2 变为 go 的template；inventory 格式 和ansible 格式是一致的。在运行方面，goansible 默认以 index.yaml 作为入口。 

### playbook

```yaml
- name: import playbook
  import_playbook: import.yaml

- name: example playbook
  hosts: test
  vars:
    d: b
  include_values:
    - values.yaml
  tasks:
    - name: echo
      shell: echo helloworld
```

上面例子描述了一playbook 结构：

* name: playbook 的名字，这是可以随便取

* import_playbook：导入另一个playbook 内容

* hosts： 主机，目前仅支持 group name

* vars: 自定义变量

* include_values：使用包含自定义变量文件，这个跟vars 合并，并重复的话会覆盖vars

* tasks： 包含一组任务


### task

```yaml
- name: shell command
  shell: echo helloword > hello
- name: copy file
  file:
    src: test.file
    dest: /tmp/test.file
- name: parse template file
  template:
    src: test.tpl
    dest: /tmp/test.tpl
```

一、变量

goansible 分为三种种变量，一种全局变量values; 一种是主机变量hostvars；还有当前组所有主机变量groupvars （hostvars 是 groupvars 一个成员）。

在模板中使用：`{{ .Values.xxx }}`、  `{{ .HostVars.xxx }}`、 `{{ .GroupVars.hostname.xxx }}`

在赋值中使用： `values.xxx`、 `hostvars.xxx` 、`groupvars.hostname.xxx`

二、循环

```yaml

- name: loop
  shell: echo {{ Item }}
  loop:
    - a
    - b
- name: loop values
  shell: echo {{ Item.xxx }}
  loop: values.loops

```

三、条件

```yaml
- name: when
  shell: echo "I am master"
  when: hostvars.k8s_master

```

四、获取执行结果

```yaml

- name: stdout
  shell: echo "stdout result"
  stdout: hostvars.stdout

```

五、打印变量

```yaml
- name: stdout
  shell: echo "stdout result"
  stdout: hostvars.stdout
  debug: {{ .hostvars.stdout }}

```

六、忽略错误

```yaml

- name: stdout
  shell: errcmd
  ignore_error: true

```

七、包含其他task文件

```yaml
- name: include task
  include: task.yaml

```

八、 标签

```yaml
- name: tag test
  shell: do something
  tag: only_do

```

运行时设置标签： 

```
./playbook -tag only_do

```

九、操作

1、shell

```yaml
- name: shell command
  shell: echo helloword > hello

- name: 使用变量模板
  shell: echo {{ .Values.test }}

```

* shell: 执行的命令，可以通过模板使用变量

2、file

```yaml
- name: copy file
  file:
    src: test.file
    dest: /tmp/test.file

```


