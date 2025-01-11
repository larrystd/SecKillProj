
启动mysql和redis服务

```bash
sudo docker-compose up
```

访问mysql和redis
```shell
mysql -uroot -h 0.0.0.0 -P23306 -proot

redis-cli -h 127.0.0.1 -p 6379 -a 123456
```

运行单元测试
```shell
go test  -run TestRegistrationScenarios SecKill/httptest
```