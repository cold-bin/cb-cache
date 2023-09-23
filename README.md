# cb-cache

## introduce

原型来自[groupcache](https://github.com/golang/groupcache/tree/master)

## implementation

- lru-k<br/>
  缓存淘汰策略支持lru-k算法（建议lru-k值取2，避免多次访问才能去除历史列表）。

> 传统lru算法有个致命的缺点：近期内大量数据被访问少量次数后，但是后面不再是热点数据，
> 可能会将近期的热点数据全部淘汰出去。只访问一次并不代表是热点数据。
> 为此，我们需要进行冷热分离，提高缓存数据的门槛，采用lru-k算法

-

## performance

## summary