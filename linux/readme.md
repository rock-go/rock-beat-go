# linux
linux下常用的数据接口

# linux.dns

监听linux模式下dns的访问记录

- userdata = linux.dns{name , region , bind}
- userdata = linux.dns(name)

#### 内部方法
- [userdata.pipe(v)]()
- [userdata.start]()
```lua
    local d = linux.dns{
        name = "monitor",
        region = region.sdk(),
        bind = "udp://0.0.0.0:53", 
    }
    d.pipe(lua.writer)
    d.pipe(function(tx)  end)
    d.start()

```