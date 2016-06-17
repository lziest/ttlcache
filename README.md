## TTLCAche - LRU cache with TTL

TTLCache is based on the implementation design of golang's groupcache lru, with cache entry TTL control. A expired
cache entry is preferred for cache eviction. When there is no expired entry, LRU principle takes effect. Cache `Get`s don't affect cache entry TTL.

It is thread-safe by a simple mutex lock.
