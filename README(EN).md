# Cah_Pool (CattleAndHorses_Pool) **`牛马`** **`Goroutine Pool`**

Cah_Pool (CattleAndHorses_Pool), also known as **`牛马`**, is a high-performance Go goroutine pool that natively supports various priority task scheduling, aiming to optimize the efficiency of concurrent task processing. It provides a simple API and powerful configuration options to flexibly adapt to different scenarios.

[中文](README.md) | [English](README(EN).md)


## Table of Contents

- [Cah_Pool (CattleAndHorses_Pool) **`牛马`** **`Goroutine Pool`**](#cah-cattleandhorses-cowhorse-goroutine-pool)
  - [Table of Contents](#table-of-contents)
  - [Description](#description)
  - [Features](#features)
  - [Performance Benchmarking](#performance-benchmarking)
  - [Installation Guide](#installation-guide)
    - [Prerequisites](#prerequisites)
    - [Installation Steps](#installation-steps)
  - [Usage Instructions](#usage-instructions)
  - [Usage Examples](#usage-examples)

## Description

- To help everyone better understand the details of this project, the project code includes **`bilingual`** comments in **`Chinese and English`**.<br>
- You are also welcome to download this project locally for testing and usage. Feedback on bugs, usage results, and test results is greatly appreciated.<br>

## Features

- **High Performance**: Utilizes **`lightweight spin locks`** and **`automatic scheduling`** logic between goroutines for efficient goroutine pool management.
- **Priority Scheduling**: Supports three types of priority task queues: **`Urgent`**, **`VIP`**, and **`Normal`**.
- **Dynamic Expansion**: Dynamically creates and recycles worker goroutines based on the waiting and running status of tasks.
- **Single Node Cluster Support**: Capable of distributing tasks across nodes, reducing internal lock contention within a single node and improving task processing efficiency.

## Performance Benchmarking
**`For reference only`** <br>
**`Benchmark code can be found in the project`** <br><br>
Testing version **`Go 1.22.6`** <br>
Testing device information:

```bash
goos: windows
goarch: amd64
cpu: AMD Ryzen 5 5600H with Radeon Graphics    
``` 
This project is compared with the popular open-source framework Ants, and the results are referenced in Tables 1-3.


 **`Benchmark Comparison Table 1`** 
| Test Item                | Operations    | Avg Time (ns/op) | Memory Allocation (B/op) | Allocations (allocs/op) | Goroutine Pool Capacity |
|---------------------|-----------|-----------------|----------------|------------------|------------|
| **Benchmark_cah**   |           |                 |                |                  | 50000      |
|          1          | 3549633   | 364.3           | 17             | 1                |            |
|          2          | 3497079   | 385.2           | 17             | 1                |            |
|          3          | 3440743   | 318.5           | 17             | 1                |            |
|          4          | 3780904   | 326.8           | 17             | 1                |            |
|          5          | 3561489   | 328.4           | 18             | 1                |            |
| **Average**            | 3565970   | 344.6           | 17             | 1                |            |
|                     |           |                 |                |                  |            |
| **Benchmark_Ants**  |           |                 |                |                  | 50000      |
|          1          | 2168454   | 466.7           | 23             | 1                |            |
|          2          | 2236189   | 471.2           | 22             | 1                |            |
|          3          | 2279091   | 513.9           | 22             | 1                |            |
|          4          | 2438413   | 465.2           | 22             | 1                |            |
|          5          | 2460843   | 513.5           | 22             | 1                |            |
| **Average**            | 2316598   | 486.1           | 22             | 1                |            |


 **`Benchmark Comparison Table 2`** 

| Goroutine Pool Capacity| Test Item             | Operation Count | Average Time(ns/op) | Memory Allocated (B/op) | Allocation Count (allocs/op) |
|------------|---------------------|----------|------------------|-----------------|----------------------|
| **10000**  | **Benchmark_cah**   | 743829   | 1608             | 18              | 1                    |
|            | **Benchmark_Ants**  | 655546   | 1626             | 23              | 1                    |
| **30000**  | **Benchmark_cah**   | 2003300  | 558.5            | 19              | 1                    |
|            | **Benchmark_Ants**  | 2213972  | 507.6            | 22              | 1                    |
| **100000** | **Benchmark_cah**   | 2583167  | 502.4            | 20              | 1                    |
|            | **Benchmark_Ants**  | 2394409  | 467.6            | 22              | 1                    |

 **`Benchmark Comparison Table 3 for Single Node Cluster (Single Node Capacity Limit 50000)`** 
| Goroutine Pool Capacity | Node Count | Test Item         | Operation Count | Average Time(ns/op) | Memory Allocated (B/op) | Allocation Count (allocs/op) |
|------------|--------|----------------------------|----------|------------------|------------------|----------------------|
| **100000** | 2      | **Benchmark_cah_Cluster**  | 3522075  | 338.9            | 19               | 1                    |
| **150000** | 3      | **Benchmark_cah_Cluster**  | 2876994  | 372.0            | 24               | 1                    |
| **200000** | 4      | **Benchmark_cah_Cluster**  | 2882150  | 358.2            | 24               | 1                    |



## Installation Guide

### Prerequisites

Make sure you have [Go](https://golang.org/dl/)installed.

### Installation Steps

Clone this repository:
```bash
git clone https://github.com/Ghostsoa/Cah_Pool.git
cd Cah
```


## Usage Instructions

Config parameters that can be specified are as follows:
- **Capacity** (Size of the goroutine pool)**`must be specified`**
- **CommonQueue** (Number of normal task queues)**`default is 3000000`**
- **VIPQueue** (Number of VIP task queues)**`default is 1000000`**
- **UrgentQueue** (Number of urgent task queues)**`default is 500`**
- **EnableCluster** (Whether to enable cluster mode)**`If enabled, please call the corresponding method`**
- **NodeCapacity** (Capacity of each node in the cluster)**`Must be specified if cluster is enabled; suggested not exceeding 100000`**

You can easily acquire a  **`singleton goroutine pool`** by:
```go
pool, err := cah.NewOnePool(config)
```
If you wish to create a **`single-node cluster`**，please use the following method:
```go
cluster, err := NewCluster(config)
```
Send tasks through  **`Submin`** . **`Submin`** accepts a func and one priority parameter; if no priority is provided, it defaults to using the normal task queue **`consistent with cluster`**:
```go
pool.Submit()
```
Supports priorities  **`1~3`** 
```go
pool.Submit(func(){},1)
```

ou can gracefully close the goroutine pool  **`consistent with cluster`** (**`Note:`** After calling this method, each goroutine will be destroyed immediately after completing its current task):

```go
pool.Close()
```

## Usage Examples

```go
package main

import (
    "fmt"
    "github.com/Ghostsoa/Cah_Pool"
)

func main() {
    // Create goroutine pool instance
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

    // Submit a normal task
    err = pool.Submit(func() {
        fmt.Println("Executing a common task.")
    })

    if err != nil {
        fmt.Printf("Failed to submit task: %v\n", err)
    }

    // Submit an urgent task
    err = pool.Submit(func() {
        fmt.Println("Executing an urgent task.")
    }, 3)

    if err != nil {
        fmt.Printf("Failed to submit urgent task: %v\n", err)
    }

    // Close the pool
    pool.Close()
}
```

