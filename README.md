# 更新 HTTPS 证书

目前用到了两个 HTTPS 证书，一个是 *.sites.interycode.com 泛域名证书，更新命令为：

```
acme.sh --issue --renew --dns -d *.sites.interycode.com --yes-I-know-dns-manual-mode-enough-go-ahead-please
cp /home/intery/.acme.sh/\*.sites.interycode.com/\*.sites.interycode.com.* /home/intery/key/
```

这个命令需要修改 DNS 的 TXT 记录，所以无法自动运行。

另一个是 interycode.com 的证书，更新命令为：

```
acme.sh --issue -d interycode.com -w /home/intery/frontend/current/
cp /home/intery/.acme.sh/interycode.com/interycode.com.* /home/intery/key/
```

上面命令可以在无人看管的情况下运行。

默认情况下，acme.sh 已经在 crontab -e 里添加了自动更新脚本，但需要修改 DNS 的时候就必须手动更新了。

所以我可能会选择为 *.sites.interycode.com 的每一个域名单独申请 HTTPS 证书，这样更自动化一点。

# 创建数据库

```
docker run --name psql1 --network=network1 -e POSTGRES_USER=intery -e POSTGRES_PASSWORD=123456 -e PGDATA=/var/lib/postgresql/data/pgdata -v intery-data-1:/var/lib/postgresql/data -d postgres
```

# 部署步骤

## 创建目录

```bash
cd ~
mkdir backend frontend log keys db config socket
```

## 创建公钥私钥

用于 jwt 加解密

```bash
ssh-keygen -t rsa -b 4096 -C <email>
```
把公钥私钥放到 `/home/intery/keys/` 下

## 设置环境变量

可以写在 ~/.bashrc 里，然后 `source ~/.bashrc`
```bash
export GIN_MODE=release

export DB_DIR=/path/to/db
export DB_HOST=127.0.0.1
export DB_USER=intery
export DB_NAME=intery_production
export DB_PASSWORD=******
export DB_PORT=5432
export PRIVATE_KEY="/home/intery/keys/id_rsa" 
export PUBLIC_KEY="/home/intery/keys/id_rsa.pub"
export NGINX_CONFIG_PATH="/home/intery/config/nginx_default.conf"
export SOCKET_DIR="/home/intery/socket"
export GITHUB_ID=
export GITHUB_SECRET=
```

## 上传 go 二进制文件

```bash
# 远程执行
mkdir /home/intery/backend
# 本地执行
CGO_ENABLED=0 go build .
scp intery intery@150.158.86.88:/home/intery/backend
```

## 启动数据库

```bash
docker run --name $DB_HOST -e POSTGRES_USER=$DB_USER -e POSTGRES_PASSWORD=$DB_PASSWORD -e PGDATA=/var/lib/postgresql/data/pgdata -v $DB_DIR:/var/lib/postgresql/data -p $DB_PORT:5432 -d postgres
cd /home/intery/backend
./intery task db:reset
```

## 启动 Nginx




# 创建全局测试覆盖率报告

在 ~/.bashrc 或者 ~/.zshrc 添加如下脚本
```
cover () {
  t="test/data/cover/go-cover.$$.tmp"
  go test -coverprofile=$t $@ && go tool cover -html=$t -o test/data/cover/cover.html && unlink $t
}
```

然后运行 `source ~/.bashrc` 或者 `source ~/.zshrc`，执行 `cover ./...` 即可