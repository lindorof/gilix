# gilix

[![Go Reference](https://pkg.go.dev/badge/github.com/lindorof/gilix.svg)](https://pkg.go.dev/github.com/lindorof/gilix)

## 介绍
基于 go 开发的「Service of Things」平台

- 通过 Acceptor 模式与外部交互，支持 tcp/http/ws 等协议
- 以 plugin 形式嵌入 Things 功能，包括语义解析、Things action 等

## 软件架构
![gilix](readme.png)

- AP 层可使用任何语言直接与 gilix 交互，只要在所支持的传输协议（tcp/http/ws/...）范围内；当然，需要针对 AP 需求来完成相匹配的 plugin 开发
- AP 若使用 C/C++ 开发，则可使用所提供的 cilix_spi_ap 库来简化访问逻辑
- 由于 Things 不可避免的需要接入 C/C++ 的 lib 库，因此提供了 RDC 封装，与 [cilix-rdc](https://github.com/lindorof/cilix-rdc) 平台远程交互
- cilix_spi_ap 和 [cilix-rdc](https://github.com/lindorof/cilix-rdc) 都通过 C 编写，合并在 [cilix-rdc](https://github.com/lindorof/cilix-rdc) 库中
- 原则：***两个一切***
    1. ***尽一切可能*** 的将复杂度放在 gilix (Go) 
    2. ***尽一切可能*** 的让 [cilix-rdc](https://github.com/lindorof/cilix-rdc) (C) 简化
- 另外，gilix-util 中提供了一些实用功能，包括：
    - 同步器 syncer
    - 打点工作器 ticker
    - 基于 zap 包装的日志库 zapt
    - ...

## How to Use

### SOT

1. ```import "github.com/lindorof/gilix"``` 

2. implement interfaces of cbs.go , such as ```Msg``` , ```Dev``` , ```Xcbs``` ...

3. ```import _ "github.com/lindorof/gilix/sot"``` to Initialize func ```NewCPS``` automatically

4. create Xcps
   
   ```go
   cps := gilix.NewCPS()
   ```

5. start the sot loop 

    ```go
    cps.SotLoopSync()
    ```

6. create acceptor as needed 

    ```go
    // ws
    import "github.com/lindorof/gilix/acp/ws"
    acceptor := ws.CreateServer(para)
    
    // tcp
    import "github.com/lindorof/gilix/acp/tcp"
    acceptor := tcp.CreateServer(para)
    
    // http
    import "github.com/lindorof/gilix/acp/http"
    acceptor := http.CreateServer(para)
    ```

7. submit acceptors to sot

    ```go
    cps.SubmitAcp(acceptor)
    ```

8. stop the sot loop on exit

    ```go
    cps.SotLoopBreak()
    ```

9.  for simplicity, recommend to use syncer

    ```go
    import "github.com/lindorof/gilix/util"
    
    // create syncer
    syncer := util.CreateSyncer(context.Background())
    
    // sync mode, returned when ctx cancelled
    syncer.Sync(
    	cps.SotLoopSync(),
    	cps.SotLoopBreak())
    
    // async mode, returned immediately
    syncer.Async(
    	cps.SotLoopSync(),
    	cps.SotLoopBreak())
    
    // for async mode, need to cancel and wait on exit
    syncer.WaitRelease(util.SyncerWaitModeCancel)
    // or just wait
    syncer.WaitRelease(util.SyncerWaitModeIdle)
    ```

### RDC

example for using of rdc

```go
import "github.com/lindorof/gilix/rdc/tcp"

type para struct {
    Name string
    Data int
}

// in/out para
in := &para{"Track1", 9}
out := &para{}

// create caller
caller := tcp.CreateCaller(":8808", 5*time.Second)
// several invocations
ret, err := caller.Invoke("ReadTrack", in, out)
// destroy caller
caller.Fini()
```

---

*lindorof . 2021.12.24* 
