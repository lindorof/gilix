# gilix

#### 介绍
基于 go 开发的「Service of Things」平台

- 通过 Acceptor 模式与外部交互，支持 tcp/http/ws 等协议
- 以 plugin 形式嵌入 Things 功能，包括语义解析、Things 功能等

#### 软件架构
![gilix](./readme.png)

- AP 层可使用任何语言直接与 gilix 交互，只要在所支持的传输协议（tcp/http/ws/...）范围内；当然，需要针对 AP 需求来完成相匹配的 plugin 开发
- AP 若使用 C/C++ 开发，则可使用所提供的 cilix_spi_ap 库来简化访问逻辑
- 由于 Things 不可避免的需要接入 C/C++ 的 lib 库，因此提供了 RDC 封装，与 cilix_rdc 平台远程交互
- cilix_spi_ap 和 cilix_rdc 都通过 C 编写，合并在 [cilix_rdc](https://gitee.com/lindorof/cilix_rdc) 库中
- 原则：**两个一切**
    1. **尽一切可能的将复杂度放在 gilix (Go) **
    2. **尽一切可能的让 [cilix_rdc](https://gitee.com/lindorof/cilix_rdc) (C) 简化**

#### 使用说明

1.  xxxx
2.  xxxx
3.  xxxx

---

*lindorof . 2021.12.24* 
