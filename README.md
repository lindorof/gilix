# gilix

[![Go Reference](https://pkg.go.dev/badge/github.com/lindorof/gilix.svg)](https://pkg.go.dev/github.com/lindorof/gilix)

## 介绍
基于 go 开发的「Service of Things」平台

- 通过 Acceptor 模式与外部交互，支持 tcp/http/ws 等协议
- 以 plugin 形式嵌入 Things 功能，包括语义解析、Things action 等

## 用途

- 适用于 WOSA/XFS/LFS/PISA/... 及其它自定义协议的 Service Provider 开发
- 基于 Go 的交叉编译特性，适用于 Windows/Linux/MacOs ，适用于 x86/x64/arm/mips 等多种 CPU 架构
- 由于 WOSA/XFS/LFS/PISA... 是基于 C 语言 的 API ，因此需要 spi 库来与 gilix 服务通讯，通讯协议由 spi 决定
- 也可以从 WEB 浏览器通过 JS 以 ws 协议的方式，直接与 gilix 服务通讯

## 语义

- 语义解析（通过 tcp/http/ws 收到的数据解析）不包含在 gilix 中
- 例如 XFS/LFS/PISA  的语义解析，属于业务开发的范畴，可使用 json（推荐）、xml、其它协议等

## 其它

- 关于 XFS/LFS/PISA 等语义解析的已有实现及其它问题可私信
- 相关 WOSA/XFS/LFS/PISA/... 业务实现过程中，会涉及其它组件，例如 Form 解析、配置处理、test 平台（GTP）等，欢迎咨询交流

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

