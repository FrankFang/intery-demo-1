# 创建数据库

```
docker run --name psql1 --network=network1 -e POSTGRES_USER=intery -e POSTGRES_PASSWORD=123456 -e PGDATA=/var/lib/postgresql/data/pgdata -v intery-data-1:/var/lib/postgresql/data -d postgres
```