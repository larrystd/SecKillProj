优惠券秒杀项目

## 版本1

### 启动

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
go test  -run TestRegistrationScenarios SecKill/test
```

### 实现原理

网络服务使用gin框架，单机单进程。缓存策略采用write back，主进程更新完缓存返回，启动后台goroutine再去更新数据库。

1. 配置路由，以/api/users开头。主要路由有三个，
    1. /api/users/:username/coupons/fetch/:name 秒杀优惠券
    2. /api/users/:username/coupons/list 列出优惠券信息，分页
    3. /:username/coupons/add 添加优惠券
2. 启动子协程执行 抢到优惠券的更新数据库逻辑
    1. 通过chan secKillChannel 接收massge信息，包括username, sellerName, couponName两个字段
    2. 从db中增加username的优惠券记录，即InsertCouponToCustomUser
    3. 减少sellerName的优惠券数量。注意sellerName下的优惠券数量是优惠券的剩余数量，而custername下的优惠券数量是custername抢到的优惠券数量

3. customer申请优惠券时，请求路由到api.FetchCoupon，执行过程
    1. 使用jwt 对请求进行鉴权, 无状态
    2. 通过redis检查coupon 剩余数量，redis通过lua脚本执行，执行逻辑
        1. 检查coupon是否有剩余，无剩余返回错误码
        2. 检查用户是否已经申请到该优惠券，申请到返错
        3. 用户拿到coupon，更新redis coupon剩余数量，用户已经拿到优惠券的信息
    3. 通过channel 将信息发给上述子协程，异步更新数据库

4. 数据库数据填充
    1. 如果mysql初始时有数据，redis-service 包在init()时会执行从mysql到redis的预热，执行时间在main函数之前；
    2. mysql redis连接的初始化也位于dao包的init()函数，
    2. 当seller 配置优惠券信息时，会同时更新db和redis。缓存需要更新的信息主要两个1. 用户持有的优惠券name，2. 每个优惠券剩余的数量

优化点
1. 服务本身是无状态的，状态记录在redis和db。 因此只需要配置负载均衡，服务可以水平扩展
2. 前台读写直接打到redis，单机的redis是性能瓶颈，可配置成分布式的。服务增加一个redis的负载均衡 
3. 为了提高流量可控性，前台服务层增加一个消息队列层，增加降级服务。
4. 前台IO虽然不直接打到DB，但需要考虑一旦缓存失效或无法服务，可能对DB的压力。
5. 需要压测给出服务能力
