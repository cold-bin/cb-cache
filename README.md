# cb-cache

## introduce
分布式的kv内存数据库

- 缓存获取流程：
  ![img.png](img.png)
  - 优先从本地cache获取kv
  - 如果没获取到，会经由一致性哈希算法找到最近的对等节点,然后获取远端节点的kv
  - 如果还没找到，可以通过`getter`设置k的数据源获取，用以缓存同步
## implementation

- `lru-k`缓存淘汰
- 多节点的一致性哈希
- 框架可扩展性
- 防止缓存击穿
- 支持并发读写