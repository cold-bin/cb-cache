# cb-cache

## introduce

原型来自[groupcache](https://github.com/golang/groupcache/tree/master)

## implementation

- 缓存淘汰
  - 采用lru-k策略，解决了传统lru可能出现的缓存命中率急剧下降的问题
- 多节点的一致性哈希
- 框架可扩展性
  - 支持自定义配置节点间通信的序列化与反序列化协议
  - 采用函数选项模式，方便框架后续扩展
- 布隆过滤器防止缓存击穿
- 支持并发读写
- 
## performance

## summary