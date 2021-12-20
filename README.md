# 创建数据库

```
docker run --name psql1 --network=network1 -e POSTGRES_USER=intery -e POSTGRES_PASSWORD=123456 -e PGDATA=/var/lib/postgresql/data/pgdata -v intery-data-1:/var/lib/postgresql/data -d postgres
```

# 创建全局测试覆盖率报告

在 ~/.bashrc 或者 ~/.zshrc 添加如下脚本
```
cover () {
  t="test/data/cover/go-cover.$$.tmp"
  go test -coverprofile=$t $@ && go tool cover -html=$t -o test/data/cover/cover.html && unlink $t
}
```

然后运行 `source ~/.bashrc` 或者 `source ~/.zshrc`，执行 `cover ./...` 即可