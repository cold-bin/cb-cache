# cb-cache

## introduce

原型来自[groupcache](https://github.com/golang/groupcache/tree/master)

## implementation

- `lru-k`缓存淘汰
- 多节点的一致性哈希
- 框架可扩展性
- 防止缓存击穿
- 支持并发读写