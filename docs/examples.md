# 示例

本目录包含 GoAnsible 的实际使用示例。

## 示例列表

### 1. 部署应用

```yaml
# playbooks/deploy.yaml
- name: deploy application
  hosts: app_servers
  tasks:
    - name: stop service
      shell: systemctl stop myapp

    - name: copy binary
      file:
        src: bin/myapp
        dest: /opt/myapp/myapp
        mode: "0755"

    - name: start service
      shell: systemctl start myapp

    - name: check health
      until:
        shell: curl -s http://localhost:8080/health
        match: "ok"
        timeout: 60
        interval: 5
```

### 2. 批量配置 Nginx

```yaml
# playbooks/nginx.yaml
- name: setup nginx
  hosts: web_servers
  vars:
    nginx_version: "1.24.0"
  tasks:
    - name: install nginx
      shell: apt-get install -y nginx={{ .Values.nginx_version }}

    - name: copy config
      template:
        src: templates/nginx.conf.tpl
        dest: /etc/nginx/nginx.conf
        mode: "0644"

    - name: copy site config
      template:
        src: templates/site.conf.tpl
        dest: /etc/nginx/sites-available/default
        mode: "0644"

    - name: restart nginx
      shell: systemctl restart nginx
```

### 3. Kubernetes 集群部署

```yaml
# playbooks/k8s.yaml
- name: deploy kubernetes
  hosts: k8s_cluster
  tasks:
    - name: setup master node
      shell: kubeadm init
      when: hostvars.node_master

    - name: join cluster
      shell: kubeadm join {{ .GroupVars.master.join_token }}
      when: not hostvars.node_master

    - name: deploy calico
      shell: kubectl apply -f calico.yaml
      when: hostvars.node_master
```

### 4. 证书管理

```yaml
# playbooks/certs.yaml
- name: generate certificates
  hosts: localhost
  tasks:
    - name: create CA
      cert:
        crt_path: certs/ca.crt
        key_path: certs/ca.key
        is_ca: "true"
        organization: myorg
        expires: 3650d

    - name: generate server cert
      cert:
        root_crt_path: certs/ca.crt
        root_key_path: certs/ca.key
        crt_path: certs/server.crt
        key_path: certs/server.key
        common_name: server.example.com
        ip_addresses: "192.168.1.100,192.168.1.101"
        expires: 365d
```

### 5. 数据库备份

```yaml
# playbooks/backup.yaml
- name: backup database
  hosts: db_servers
  tasks:
    - name: create backup dir
      directory:
        path: /backup/mysql
        mode: "0755"

    - name: dump database
      shell: mysqldump -u root {{ .Values.db_name }} > /backup/mysql/{{ .Values.db_name }}_{{ .Values.backup_date }}.sql

    - name: compress backup
      shell: gzip /backup/mysql/{{ .Values.db_name }}_{{ .Values.backup_date }}.sql

    - name: upload to s3
      shell: aws s3 cp /backup/mysql/ s3://{{ .Values.s3_bucket }}/mysql/ --recursive
```

### 6. 监控部署

```yaml
# playbooks/monitoring.yaml
- name: setup monitoring
  hosts: monitoring_servers
  tasks:
    - name: install prometheus
      shell: |
        wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz
        tar xvf prometheus-*.tar.gz
        mv prometheus-* /opt/prometheus

    - name: copy config
      template:
        src: templates/prometheus.yml.tpl
        dest: /opt/prometheus/prometheus.yml

    - name: install grafana
      shell: |
        wget https://dl.grafana.com/oss/release/grafana_10.0.0_amd64.deb
        dpkg -i grafana_*.deb

    - name: start services
      shell: |
        systemctl start prometheus
        systemctl start grafana-server
```

### 7. 使用自定义模块

```yaml
# playbooks/custom.yaml
- name: use custom modules
  hosts: all
  tasks:
    - name: generate random password
      genpass:
        length: 32
      setface: "values.password=ctx.res"

    - name: parse version
      parsever:
        input: "1.2.3"
      setface: "values.major=ctx.major"
      setface: "values.minor=ctx.minor"
      setface: "values.patch=ctx.patch"
```

### 8. CI/CD 流水线

```yaml
# playbooks/cicd.yaml
- name: CI/CD pipeline
  hosts: ci_server
  tasks:
    - name: checkout code
      shell: git clone {{ .Values.repo_url }} /workspace/app

    - name: run tests
      shell: cd /workspace/app && go test ./...

    - name: build
      shell: cd /workspace/app && go build -o bin/app

    - name: deploy to staging
      import_playbook: deploy.yaml
      tags: staging

    - name: run smoke tests
      shell: curl -f http://staging.example.com/health

    - name: deploy to production
      import_playbook: deploy.yaml
      when: values.approved
      tags: production
```

---

## 模板文件示例

### nginx.conf.tpl

```
worker_processes auto;

events {
    worker_connections {{ .Values.worker_connections | default 1024 }};
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    sendfile on;
    keepalive_timeout 65;

    upstream app {
        {{- range .Values.backend_servers }}
        server {{ . }};
        {{- end }}
    }

    server {
        listen 80;
        server_name {{ .Values.server_name }};

        location / {
            proxy_pass http://app;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
```

### values.yaml 示例

```yaml
# 应用配置
app:
  name: myapp
  port: 8080
  replicas: 3

# SSH 配置
ssh:
  user: deploy
  port: 22

# 后端服务器
backend_servers:
  - 192.168.1.10:8080
  - 192.168.1.11:8080
  - 192.168.1.12:8080

# 服务器配置
server_name: example.com
worker_connections: 2048

# 数据库
db:
  host: 192.168.1.20
  port: 5432
  name: myapp
  user: admin
```
