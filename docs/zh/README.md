# EDS

**EDS**（Edge Device SDK）是为 IOT 场景设计并开发的设备接入层 SDK，提供快速将多种协议的设备接入平台的能力。

## 特性

1. 轻量：只使用 MQTT 作为设备服务与设备中心之间数据交换的中间件，无需引入多余组件；
2. 通用：
    1. 将 MQTT 封装为 MessageBus，向上层模块提供支持；
    2. 基于 MessageBus 定义和实现了通用的**元数据操作**与**物模型**操作规范。

## 术语

### 物模型

[物模型](https://blog.csdn.net/zjccoder/article/details/107050046)，作为一类设备的抽象，描述了设备的：

- 属性（property）：用于描述设备状态，支持读取和写入；
- 方法（method）：设备可被外部调用的能力或方法，可设置输入和输出参数，参数必须是某个“属性”，相比与属性，方法可通过一条指令实现更复杂的业务逻辑；
- 事件（event）：用于描述设备主动上报的事件，可包含多个输入参数，参数必须是某个“属性”。

产品（Product）：即物模型。

设备（Device）：与真实设备一一对应，必须属于并且仅属于某一个 Product。

### 元数据

元数据包括协议（Protocol）、产品（Product）及设备（Device）：

- 协议对应了一个特定协议的描述信息；
- 产品对应了一个特定产品的描述信息；
- 设备对应了一个特定设备的描述信息。

### 设备服务

设备协议驱动服务的简称，负责一种特定的设备协议的接入实现。主要功能包括：

- 协议注册：将当前设备协议注册到设备管理中心；
- 设备初始化：从设备重新拉取产品 & 设备等元数据，加载驱动；
- 元数据监听：监听设备中心产品 & 设备等元数据的变更，加载/卸载/重加载驱动；
- 数据接入：与实际设备进行连接，向 MessageBus 发送采集的设备数据（设备中心负责从 MessageBus 中读取出数据）;
- 命令执行：与实际设备进行连接，从 MessageBus 获取调用的方法执行（设备中心负责将执行命令写入到 MessageBus）。

### 设备中心

设备管理中心的简称，由 SDK 提供基础服务，嵌入到平台中使用。主要功能包括：

- 协议管理：接收设备服务的注册请求 & 设备服务探活（失活则更新设备元数据）；
- 产品管理：基于特定协议定义产品 & 产品 CRUD；
- 设备管理：基于特定产品定义设备 & 设备 CRUD。

## Topic 约定

因为 EDS 基于 MQTT 实现数据交换，所以我们基于 MQTT 的 Topic & Payload 概念定义了我们自己的数据格式及通信规范。

### 物模型

对于物模型来说，Topic 格式为 `DATA/{ProductID}/{DeviceID}/{OptType}/{DataID}`：

- `ProductID`：设备元数据中产品的 UUID，产品唯一；
- `DeviceID`：设备元数据中设备的 UUID，设备唯一；
- `OptType`，产生当前数据的操作类型：
    - 对于 `property`，可选 `read | write`，分别对应于属性的读取和写入；
    - 对于 `method`，可选 `request | response | error`，分别对应于方法的请求、响应和出错；
    - 对于 `event`，可选 `event`，分别对应于事件上报。
- `DataID`，当前数据的 UUID：
    - 对于隶属于同一个方法调用 `method call`，其 `request` 与 `response` 的 `DataID` 是相同的；
    - 对于非方法调用的其它类型的数据，其 `DataID` 都是不同的。

### 元数据操作

对于元数据增删改查等操作来说，Topic 格式为 `META/{MetaType}/{Method}/{OptType}/{DataID}`：

- `MetaType`：元数据类型，可选 `protocol | product | device`；
- `MethodType`：调用方法，可选 `create | update | delete | get | list`，对于不同的元数据类型，可选范围是不同的；
- `DataType`：数据类型，可选 `request | response | error`；
- `DataID`：当前数据的 UUID，特别地，对于同一个方法，其 `request` 与 `response` 的 `DataID` 是相同的。