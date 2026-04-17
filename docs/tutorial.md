# GoAnsible 使用教程

本教程详细介绍 GoAnsible 的各项功能和使用方法。

## 目录

1. [Inventory 配置](#inventory-配置)
2. [Playbook 结构](#playbook-结构)
3. [Task 操作](#task-操作)
4. [变量和函数](#变量和函数)
5. [高级功能](#高级功能)
6. [自定义模块](#自定义模块)

---

## Inventory 配置

Inventory 文件使用 YAML 格式，定义主机属性和分组。

### 基本结构

```yaml
groups:
  all: # 必须定义，包含所有主机的定义
    host1:
      ansible_ssh_host: "192.168.1.100"
      ansible_ssh_port: "22"
      ansible_ssh_user: root
      ansible_ssh_pass: "123456"
      ansible_ssh_sudopass: "123456"  # sudo 密码（可选）
      ansible_ssh_key: ""  # SSH 密钥（与密码二选一）
    host2:
      ansible_ssh_host: "192.168.1.101"
      ansible_ssh_port: "22"
      ansible_ssh_user: root
      ansible_ssh_pass: "123456"
    localhost: # 本地主机，无需 SSH
      ansible_ssh_host: ""
      ansible_ssh_port: ""
      ansible_ssh_user: ""
      ansible_ssh_pass: ""
      ansible_ssh_key: ""

  testgroup1: # 自定义组
    host1:
      node_master: true
    host2: {}

vars:
  key1: value1
  app_name: myapp
```

### 字段说明

| 字段 | 说明 |
|------|------|
| `ansible_ssh_host` | 主机 SSH IP 地址 |
| `ansible_ssh_port` | 主机 SSH 端口，默认 22 |
| `ansible_ssh_user` | 主机 SSH 用户名 |
| `ansible_ssh_pass` | 主机 SSH 密码 |
| `ansible_ssh_key` | 主机 SSH 密钥路径 |
| `ansible_ssh_sudopass` | sudo 密码，设置后所有命令自动加 sudo |

---

## Playbook 结构

### 基本结构

```yaml
- name: example playbook
  hosts: testgroup1
  vars:
    d: b
  include_values:
    - values.yaml
  tasks:
    - name: echo
      shell: echo helloworld
```

### 字段说明

| 字段 | 说明 |
|------|------|
| `name` | Playbook 名称 |
| `import_playbook` | 导入其他 playbook |
| `hosts` | 目标主机组 |
| `vars` | 自定义变量 |
| `include_values` | 包含变量文件 |
| `tasks` | 任务列表 |
| `tags` | 标签，用于过滤执行 |

### 导入 Playbook

```yaml
- name: import playbook
  import_playbook: import.yaml
  tags:
    - tag1
    - tag2
```

---

## Task 操作

### 1. Shell 命令

```yaml
- name: shell command
  shell: echo helloworld > /tmp/hello

- name: 使用变量
  shell: echo {{ .Values.test }}
```

### 2. 文件操作

#### File - 复制文件

```yaml
- name: copy file
  file:
    src: test.file
    dest: /tmp/test.file
    mode: "0644"
    owner: root
    group: root
```

#### Template - 模板文件

```yaml
- name: parse template file
  template:
    src: test.tpl
    dest: /tmp/test.tpl
    mode: "0755"
    owner: root
    group: root
```

#### Directory - 创建目录

```yaml
- name: create directory
  directory:
    path: /tmp/test
    mode: "0755"
    owner: root
    group: root
```

### 3. 正则匹配

```yaml
- name: regexp match
  regexp:
    src: abrrrc
    exp: a(b{1})r(r{2})c
    dst: values.reg.result

- name: regexp split
  regexp:
    src: "a,b,c"
    exp: ","
    dst: values.arr
    split: true
```

### 4. Until - 循环等待

```yaml
- name: wait service
  until:
    shell: "ss -lnp|awk '{print $5}'|grep -n :3000$"
    match: ".+"
    timeout: 300
    interval: 10
```

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `shell` | 执行的命令 | - |
| `match` | 匹配正则表达式 | - |
| `timeout` | 超时时间（秒） | 300 |
| `interval` | 执行间隔（秒） | 5 |

### 5. 证书生成

```yaml
- name: 生成 CA
  cert:
    crt_path: root.crt
    key_path: root.key
    is_ca: "true"
    country: CN
    organization: k8s
    organizational_unit: opsnull
    expires: 100y

- name: 生成证书
  cert:
    root_crt_path: root.crt
    root_key_path: root.key
    crt_path: test.crt
    key_path: test.key
    is_ca: "false"
    common_name: kubernetes-master
    country: CN
    organization: k8s
    expires: 100y
    ip_addresses: "127.0.0.1,192.168.1.1"
    domains: "localhost"
```

### 6. Curl 请求

```yaml
- name: POST request
  curl:
    url: http://127.0.0.1/api
    option:
      request: POST
      header: "Content-Type:application/json"
      data: |
        {"hello":"world"}

- name: YAML to JSON
  curl:
    url: http://127.0.0.1/api
    option:
      request: POST
      header: "Content-Type:application/json"
      data-yaml: |
        hello: world
```

### 7. 设置变量

```yaml
- name: set variable
  setface: 'values.a=hostvars.ansible_ssh_host'
  debug: '{{ .Values.a }}'
```

### 8. 脚本支持

#### JavaScript

```yaml
- name: js script
  js: |
    var v = 'js value'
    values.key1 = v
  debug: '{{ .Values.key1 }}'
```

#### Lua

```yaml
- name: lua script
  lua: |
    local v = "lua value"
    values.key1 = v
  debug: '{{ .Values.key1 }}'
```

---

## 变量和函数

### 变量类型

GoAnsible 支持四种变量：

| 变量类型 | 说明 | 访问方式 |
|----------|------|----------|
| `values` | 全局变量 | `{{ .Values.xxx }}` |
| `hostvars` | 主机变量 | `{{ .HostVars.xxx }}` |
| `groupvars` | 组变量 | `{{ .GroupVars.hostname.xxx }}` |
| `groups` | 所有组变量 | `{{ .Groups.groupname.hostname.xxx }}` |

### 赋值使用

在 `when`, `loop`, `stdout` 等字段中使用：

- `values.xxx`
- `hostvars.xxx`
- `groupvars.hostname.xxx`
- `groups.groupname.hostname.xxx`

### 内置函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `join` | 数组合并为字符串 | `{{ join .Values.testArray "," }}` |
| `join_groupvars` | 组变量合并 | `{{ join_groupvars .GroupVars ansible_ssh_host ","}}` |
| `plus` | 加法 | `{{ plus .Values.a .Values.b }}` |
| `minus` | 减法 | `{{ minus .Values.a .Values.b }}` |

还支持 [Sprig](http://masterminds.github.io/sprig/) 第三方函数库。

---

## 高级功能

### 循环

```yaml
- name: loop items
  shell: echo {{ .Item }}
  loop:
    - a
    - b

- name: loop with concurrency
  shell: echo {{ .Item.xxx }}
  concurrency: 2
  loop: values.loops
```

### 条件

```yaml
- name: conditional task
  shell: echo "I am master"
  when: hostvars.k8s_master
```

### 获取执行结果

```yaml
- name: capture output
  shell: echo "result"
  stdout: hostvars.stdout
  debug: "{{ .HostVars.stdout }}"
```

### 忽略错误

```yaml
- name: ignore error
  shell: errcmd
  ignore_error: true
```

### 包含 Task 文件

```yaml
- name: include tasks
  include: task.yaml
```

### 标签

```yaml
- name: tagged task
  shell: do something
  tags:
    - tag1
    - tag2
```

运行时使用标签过滤：

```bash
./goansible -tags tag1,tag2
```

---

## 自定义模块

### YAML 模块

创建 `modules/ls.yaml`:

```yaml
- name: list dir
  shell: ls {{ .Ctx.dir }}
```

使用:

```yaml
- name: ls
  ls:
    dir: /tmp
```

### JavaScript 模块

创建 `modules/split.js`:

```js
ctx.res = ctx.data.split(',')
```

使用:

```yaml
- name: split string
  split:
    data: "1,2,3"
  setface: "values.arr=ctx.res"
  debug: "{{ .Values.arr }}"
```

### JavaScript 库

创建 `libs/strings.js`:

```js
class Strings {
  random(e) {
      e = e || 32;
      var t = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678",
      a = t.length,
      n = "";
      for (i = 0; i < e; i++) n += t.charAt(Math.floor(Math.random() * a));
      return n
  }
}

var strings = new Strings()
```

使用:

```yaml
- name: generate random
  js: |
    values.a = strings.random(32)
  debug: "{{ .Values.a }}"
```

---

## 最佳实践

1. **变量管理**: 将敏感信息放在 `values.yaml` 中，不要提交到版本控制
2. **模块化**: 将常用操作封装成自定义模块
3. **标签**: 为任务添加标签便于选择性执行
4. **错误处理**: 对可能失败的任务使用 `ignore_error`
5. **并发**: 对批量操作使用 `concurrency` 控制并发数
