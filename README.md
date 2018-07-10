# go-elasticache

This is a fork of [Integralist/go-elasticache](https://github.com/Integralist/go-elasticache) with a few changes and additions:

- [Watcher support](#auto-detect-cluster-changes) for detecting and applying server changes.
- Removed logging and reliance on environment variables.
- Remove unnecessary `Node` and `Item` types.
- Replaced `glide` with `dep` as dependency manager.
- Moved package to `go-elasticache` instead of `go-elasticache/elasticache`.

## Auto Detect Cluster Changes

You can use the `Watch` method to detect changes in cluster configuration. The client is updated automatically when changes are detected.

```go
// Make sure the endpoint has ".cfg" in it!
client, _ := elasticache.New("mycluster.fnjyzo.cfg.use1.cache.amazonaws.com")

ctx := context.WithCancel(context.Background())
defer cancel() // Call cancel when you are done with the memcache client

go client.Watch(ctx)
```

See [Elasticache Auto Discovery](https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/AutoDiscovery.html) docs for details.