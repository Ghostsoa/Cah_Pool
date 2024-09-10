# cah  **`全称`**( CattleAndHorses) **`牛马`** **`协程池`**

cah(CattleAndHorses)，中文名 **`牛马`** 是一个高性能的 Go 协程池，原生支持多种优先级任务调度，旨在优化并发任务的处理效率。提供了简单的 API 和强大的配置选项，以便灵活地应用于不同场景。

## 目录

- [cah **`全称`** (  CattleAndHorses) **`牛马`** **`协程池`**](#cah--全称-cattleandhorses-牛马-协程池)
  - [目录](#目录)
  - [说明](#说明)
  - [功能特性](#功能特性)
  - [性能基准测试](#性能基准测试)
  - [安装指南](#安装指南)
    - [前提条件](#前提条件)
    - [安装步骤](#安装步骤)
  - [用法说明](#用法说明)
  - [用法示例](#用法示例)

## 说明
- 为了保证各位更好的了解本项目的细节，项目代码使用了 **`中英`** 双语、详细的注释。<br>
- 同时，也欢迎各位下载本项目到本地进行测试和使用。欢迎各位提交bug、使用结果、测试结果的反馈<br>

## 功能特性

- **高性能**: 使用 **`轻量化自旋锁`** 和 **`协程间`** 自动调度逻辑，以实现高效的协程池管理。
- **优先级调度**: 支持 **`紧急`**、**`优先`** 和 **`普通`** 三种优先级任务队列。
- **动态扩展**: 根据任务的等待和运行情况动态创建和回收工作协程。
- **单机集群支持**: 能够在节点之间分配任务，减少单节点内部锁的竞争，提高任务处理效率。
  
## 性能基准测试
**`仅供参考`** <br>
**`测试代码可在项目中查看`** <br><br>
测试设备信息<br> 测试版本 **`go 1.22.6`**
```bash
goos: windows
goarch: amd64
cpu: AMD Ryzen 5 5600H with Radeon Graphics     
``` 
本项目与目前主流的开源框架Ants进行对比,结果参考对比表1~3 <br>

 **`基准测试对比表1`** 
| 测试项目                | 操作次数    | 平均时间(ns/op) | 内存分配(B/op) | 申请次数(allocs/op) | 协程池容量 |
|---------------------|-----------|-----------------|----------------|------------------|------------|
| **Benchmark_cah**   |           |                 |                |                  | 50000      |
|          1          | 3549633   | 364.3           | 17             | 1                |            |
|          2          | 3497079   | 385.2           | 17             | 1                |            |
|          3          | 3440743   | 318.5           | 17             | 1                |            |
|          4          | 3780904   | 326.8           | 17             | 1                |            |
|          5          | 3561489   | 328.4           | 18             | 1                |            |
| **平均**            | 3565970   | 344.6           | 17             | 1                |            |
|                     |           |                 |                |                  |            |
| **Benchmark_Ants**  |           |                 |                |                  | 50000      |
|          1          | 2168454   | 466.7           | 23             | 1                |            |
|          2          | 2236189   | 471.2           | 22             | 1                |            |
|          3          | 2279091   | 513.9           | 22             | 1                |            |
|          4          | 2438413   | 465.2           | 22             | 1                |            |
|          5          | 2460843   | 513.5           | 22             | 1                |            |
| **平均**            | 2316598   | 486.1           | 22             | 1                |            |


 **`基准测试对比表2`** 

| 协程池容量 | 测试项目                | 操作次数 | 平均时间(ns/op) | 内存分配(B/op) | 申请次数(allocs/op) |
|------------|---------------------|----------|------------------|-----------------|----------------------|
| **10000**  | **Benchmark_cah**   | 743829   | 1608             | 18              | 1                    |
|            | **Benchmark_Ants**  | 655546   | 1626             | 23              | 1                    |
| **30000**  | **Benchmark_cah**   | 2003300  | 558.5            | 19              | 1                    |
|            | **Benchmark_Ants**  | 2213972  | 507.6            | 22              | 1                    |
| **100000** | **Benchmark_cah**   | 2583167  | 502.4            | 20              | 1                    |
|            | **Benchmark_Ants**  | 2394409  | 467.6            | 22              | 1                    |

 **`rah单机集群(单节点上线容量50000)基准测试对比表3`** 
| 协程池容量 | 节点数 | 测试项目                   | 操作次数 | 平均时间(ns/op) | 内存分配(B/op) | 申请次数(allocs/op) |
|------------|--------|----------------------------|----------|------------------|------------------|----------------------|
| **100000** | 2      | **Benchmark_cah_Cluster**  | 3522075  | 338.9            | 19               | 1                    |
| **150000** | 3      | **Benchmark_cah_Cluster**  | 2876994  | 372.0            | 24               | 1                    |
| **200000** | 4      | **Benchmark_cah_Cluster**  | 2882150  | 358.2            | 24               | 1                    |



## 安装指南

### 前提条件

确保已安装 [Go](https://golang.org/dl/)。

### 安装步骤

1. 克隆此仓库：
    ```bash
    git clone https://github.com/Ghostsoa/Cah.git
    cd cah
    ```

2. 下载依赖：
    ```bash
    go mod tidy
    ```

## 用法说明

可指定的Config参数如下:
- **Capacity** (协程池中的工作池大小)**`必须指定`**
- **CommonQueue** (普通任务队列数)**`默认3000000`**
- **VIPQueue** (VIP 任务队列数)**`默认1000000`**
- **UrgentQueue** (紧急任务队列数)**`默认500`**
- **EnableCluster** (是否开启集群模式)**`如果开启，请调用对应方法`**
- **NodeCapacity** (集群中单个节点的容量)**`如果开启集群，必须指定，建议不要超过10w`**

你可以简单的通过以下方式获取一个 **`单例协程池`** :
```go
pool, err := cah.NewOnePool(config)
```
如果你希望创建 **`单机集群`**，请使用以下方法:
```go
cluster, err := NewCluster(config)
```
通过 **`Submin`** 发送任务， **`Submin`** 接收一个func函数，与1个优先级参数，不传递优先级默认为使用普通任务队列 **`集群一致`**
```go
pool.Submit()
```
支持 **`1~3`** 的优先级
```go
pool.Submit(func(){},1)
```

你可以优雅的关闭协程池 **`集群一致`** (**`注:`** 调用该方法后，每个协程会在完成当前任务后，直接销毁)
```go
pool.Close()
```

## 用法示例

```go
package main

import (
    "fmt"
    "github.com/yourusername/cattleandhorses"
)

func main() {
    // 创建协程池实例
    config := cah.Config{
        Capacity:      10,
        CommonQueue:   100,
        VIPQueue:      50,
        UrgentQueue:   20,
        EnableCluster: false,
    }
    
    pool, err := cah.NewOnePool(config)
    if err != nil {
        panic(err)
    }

    // 提交一个普通任务
    err = pool.Submit(func() {
        fmt.Println("Executing a common task.")
    })

    if err != nil {
        fmt.Printf("Failed to submit task: %v\n", err)
    }

    // 提交一个紧急任务
    err = pool.Submit(func() {
        fmt.Println("Executing an urgent task.")
    }, 3)

    if err != nil {
        fmt.Printf("Failed to submit urgent task: %v\n", err)
    }

    // 关闭池
    pool.Close()
}
```
