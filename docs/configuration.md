# 配置参考

本文档详细说明 GoAnsible 的配置文件格式和选项。

## 配置文件

### values.yaml

全局变量配置文件：

```yaml
# SSH 配置
ssh:
  user: root
  port: 22
  timeout: 30

# 应用配置
app:
  name: myapp
  version: 1.0.0
  port: 8080

# 环境配置
env:
  production: true
  debug: false

# 主机列表
hosts:
  - 192.168.1.100
  - 192.168.1.101
  - 192.168.1.102
```

### index.yaml

Playbook 索引文件：

```yaml
playbooks:
  - name: deploy
    file: playbooks/deploy.yaml
  - name: backup
    file: playbooks/backup.yaml
```

---

## Playbook 配置项

### 顶层字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | Playbook 名称 |
| `hosts` | string | 是 | 目标主机组名 |
| `vars` | object | 否 | 变量定义 |
| `include_values` | list | 否 | 包含的变量文件列表 |
| `tasks` | list | 是 | 任务列表 |
| `import_playbook` | string | 否 | 导入的 playbook 文件 |
| `tags` | list | 否 | 标签列表 |

---

## Task 配置项

### 通用字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 任务名称 |
| `when` | string | 执行条件 |
| `loop` | list/string | 循环配置 |
| `concurrency` | int | 并发数 |
| `tags` | list | 任务标签 |
| `ignore_error` | bool | 忽略错误 |
| `debug` | string | 调试输出 |
| `stdout` | string | 输出保存变量 |

### 操作类型

#### shell

```yaml
shell: <command string>
```

#### file

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `src` | string | - | 源文件路径 |
| `dest` | string | - | 目标路径 |
| `mode` | string | 0644 | 文件权限 |
| `owner` | string | ssh user | 文件所有者 |
| `group` | string | ssh group | 文件组 |

#### template

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `src` | string | - | 模板文件路径 |
| `dest` | string | - | 目标路径 |
| `mode` | string | 0644 | 文件权限 |
| `owner` | string | ssh user | 文件所有者 |
| `group` | string | ssh group | 文件组 |

#### directory

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `path` | string | - | 目录路径 |
| `mode` | string | 0755 | 目录权限 |
| `owner` | string | ssh user | 所有者 |
| `group` | string | ssh group | 组 |

#### regexp

| 字段 | 类型 | 说明 |
|------|------|------|
| `src` | string | 源字符串 |
| `exp` | string | 正则表达式 |
| `dst` | string | 结果保存变量 |
| `split` | bool | 是否分割 |

#### until

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `shell` | string | - | 执行的命令 |
| `match` | string | - | 匹配正则 |
| `timeout` | int | 300 | 超时时间（秒） |
| `interval` | int | 5 | 执行间隔（秒） |

#### cert

| 字段 | 类型 | 说明 |
|------|------|------|
| `crt_path` | string | 证书输出路径 |
| `key_path` | string | 密钥输出路径 |
| `root_crt_path` | string | CA 证书路径（签发证书时） |
| `root_key_path` | string | CA 密钥路径（签发证书时） |
| `is_ca` | string | 是否为 CA |
| `common_name` | string | 通用名 |
| `country` | string | 国家 |
| `organization` | string | 组织 |
| `organizational_unit` | string | 组织单位 |
| `expires` | string | 有效期，如 100y |
| `ip_addresses` | string | IP 地址列表，逗号分隔 |
| `domains` | string | 域名列表，逗号分隔 |

#### curl

| 字段 | 类型 | 说明 |
|------|------|------|
| `url` | string | 请求 URL |
| `option.request` | string | 请求方法 |
| `option.header` | string | 请求头 |
| `option.data` | string | 请求数据（JSON） |
| `option.data-yaml` | string | 请求数据（YAML） |

#### js

```yaml
js: |
  <javascript code>
```

可用的上下文对象：
- `values` - 全局变量
- `hostvars` - 主机变量
- `groupvars` - 组变量
- `groups` - 所有组

#### lua

```yaml
lua: |
  <lua code>
```

可用的上下文对象同 js。

---

## 环境变量

可以通过环境变量覆盖配置：

```bash
export GOANSIBLE_SSH_USER=admin
export GOANSIBLE_SSH_PORT=2222
```
