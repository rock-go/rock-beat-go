# windows
windows下的信息采集接口 主要包括eventlog、registtry、wmi的api

# win.event
- ud = win.event{name , begin , pipe , pass}
- name: 服务名称
- begin: 是否强制开始区读取 
- pipe：事件的处理逻辑     
- pass: 不处理的事件     
#### 函数接口
- [ud.to(lua.writer)]()
- [ud.pipe(pipe)]()
- [ud.start()]()

#### event 字段
- [ev.xml]()
- [ev.provider_name]()
- [ev.event_id]()
- [ev.task]()
- [ev.op_code]()
- [ev.create_time]()
- [ev.record_id]()
- [ev.process_id]()
- [ev.thread_id]()
- [ev.channel]()
- [ev.computer]()
- [ev.version]()
- [ev.render_field_err]()
- [ev.message]()
- [ev.level_text]()
- [ev.task_text]()
- [ev.op_code_text]()
- [ev.keywords]()
- [ev.channel_text]()
- [ev.id_text]()
- [ev.publish_err]()
- [ev.bookmake]()
- [ev.subscribe]()
- [ev.exdata]()
- [ev.Json()]()

```lua
    local wev = win.event{
       name = "login",
       begin = false,
    } 


   wev.to(lua.Writer)     -- 添加事件收集
   wev.pipe(function(ev) end)    -- 添加多个事件处理器
   wev.start()


   wev.ev_1001 = function(ev) --event id 1001 handle
    --ev.op_code
   end
```

#### eventdata 

- 满足index和newindex 字段根据eventlog xmlEvent字段来
- 比如登录事件的eventdata 事件ID: 4624
```lua
    local exdata = ev.exdata()

    print(exdata.op_code)
    print(exdata.channel)
    print(exdata.proccessId)

    print(exdata.IpAddress)
    print(exdata.TargetUserName)
    print(exdata.TargetUserSid)
    print(exdata.SubjectLogonId)
    print(exdata.IpPort)
    print(exdata.ProcessName)
    print(exdata.vVirtualAccount)
    --具体的查看官方手册 或者 windows事件日志
```

# win.registry.*

操作windows的注册表
- [win.registry.USERS]()
- [win.registry.LOCAL_MACHINE]()
- [win.registry.CLASSES_ROOT]()
- [win.registry.CURRENT_CONFIG]()
- [win.registry.CURRENT_USER]()
- [win.registry.PERFORMANCE_DATA]()
 
## 使用 
- 操作和获取的接口
- [set(key , perm)](set)
- [open(key)](open)
- [pipe(key , function)](pipe)
```lua
    local item , exist , err= win.registry.CURRENT_USER.set("\\SOFTWARE\\Tencent\\WeChat")
    local item = win.registry.CURRENT_USER.open("\\SOFTWARE\\Tencent\\WeChat")
```
#### set
- item , exist , err = win.registry.*.set(key , perm)
- key: 路径
- perm: 权限设置
- [ALL_ACCESS]() 
- [CREATE_LINK]() 
- [CREATE_SUB_KEY]() 
- [EXECUTE]() 
- [NOTIFY]()
- [QUERY_VALUE]()
- [READ]()
- [SET_VALUE]()
- [WOW65_32KEY]()
- [wow65_64KEY]()
- [WRITE]()
- [ENUMERATE_SUB_KEYS]()
-
#### open
- item = win.registry.*.set(key)
- key: 路径

#### pipe
- error = win.registry.*.pipe(key , function)
- 遍历所有的值
- [item.prefix]()
- [item.name]()
- [item.path]()
- [item.value]()
```lua
    local err = win.registry.CURRENT_USER.pipe("\\SOFTWARE\\Tencent\\WeChat" , function(item)
        print(item.prefix)
        print(item.name)
        print(item.path)
        print(item.value)
    end)

```

#### 键值方法 
- 单个item值的操作接口
- [item.s(string)]()
- [item.n(string)]()
- [item.b(string)]()
- [item.get(string)]()
- [item.keys()]()
- [item.count()]()
- [item.sub_keys()]()
- [item.sub_count()]()
- [item.map()]()
- [item.close()]()
- [item.set_qword()]()
- [item.set_dwrod()]()
- [item.set_strings()]()
- [item.set_expand()]()

```lua
    local cu = win.registry.CURRENT_USER
    
    local chat = cu.open("\\SOFTWARE\\Tencent\\WeChat")

    print(chat.s'FileSavePath') --获取FileSavePath string 类型
    print(chat.n'LANG_ID')      --获取LANG_ID      Int    类型
    print(chat.b'xxxx')         --获取xxxx         Bool   类型

    print(chat.get'xxxx')       --获取xxxx 
    
    local keys = chat.sub_keys() --子键
    local count = chat.sub_count() --数量

    local keys = chat.keys()       -- keys
    local count = chat.count()     -- 数量

    local chat , ex , err = cu.set("\\SOFTWARE\\Tencent\\WeChat" , "ALL_ACCESS")
    print(ex)  --是否已经存在了

    local item = chat.get(FileSavePath)
    print(item.set_qword(10))
    print(item.set_dword(10))
    print(item.strings{"aa" , "bb"})
    print(item.expand("aaaa"))

    for key , val in cu.map() do
        print(key)
        print(val)
    end 
```

# win.wmi.*

- winddows wmi 底层
- [win.wmi.query(sql , function)]()

#### item 接口
- [item.s(string)]()
- [item.n(string)]()
- [item.b(string)]()
- [item.t(string)]()
```lua
    win.wmi.query("SELECT * FROM Win32_Account" , function(item)
        print(item.s'Name')
        print(item.n'AccountType' or -1)
        print(item.s'Caption')
        print(item.s'Description')
        print(item.b'Disabled')
        print(item.s'Domain')
        print(item.s'FullName')
        print(item.s'InstallDate')
        print(item.b'LocalAccount')
        print(item.b'Lockout')
        print(item.b'PasswordChangeable')
        print(item.b'PasswordExpires')
        print(item.b'PasswordRequired')
        print(item.s'SID')
        print(item.s'SIDType')
        print(item.s'Status') 
    end)
```