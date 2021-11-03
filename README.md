# DEHUS
**DEBUS** (Device Bus的简称)，属于设备接入模块，很多其他公司称其为IOT HUB。由于设备的协议区别很大，但是为了更好地支持设备与设备间，设备与服务器通信，我们通过DEHUB来完成协议间的转换以及通信。

> #### Debus的功能：
> * 支持多种设备接入协议：支持设备使用MQTT、OPC-UA协议接入物联网平台。
> * 支持多种通信模式：支持RPC和PUB/SUB两种通信模式。

## 第一期目标：
* 设备mqtt的pub的消息，通过DEBUS能传到所有sub这个消息的客户端。
* pub/sub消息的topic的定义，如下：
    * 每个DEVICE我们都会分配唯一的uuid
    * 除此之外，每个DEVICE唯一地属于某一类产品product
    * 于是/${product}/${uuid}构成了DEVICE的唯一标识
* 设备端的pub的topic规范：
    * /${product}/${uuid}/output
    * /${product}/${uuid}/output/error
* 设备端的sub的topic规范：
    * /${product}/${uuid}/input
* 附上MQTT的协议 [链接](http://public.dhe.ibm.com/software/dw/webservices/ws-mqtt/mqtt-v3r1.html)
* 设备端sub/pub的消息体body规范：
    * json格式
    * 将数据以key-value形式保存
    * 样例：{ "temperature": 31.6, "humidity": 120, "pm25": 56 }

## 第二期目标
* 支持更多协议OPC-UA
* 我们可能会扩充一些系统设备的规范, 如下
* 服务端的pub/sub规范：
    * /sys/${product}/${uuid}/std/${custom_name}
* 服务端的rpc规范：
    * /sys/${product}/${uuid}/rpc/req/${msg_uuid}
    * /sys/${product}/${uuid}/rpc/rsp/${msg_uuid}

## Topic约定
ThingerDebus模块处理的mqtt topic主要有以下几类:

* 系统通知主题, 使用统一数据结构
  ALERT/{instance_id}
    - instance_id 消息所属实例ID
    - 构建方法: mqtt.NewAlertTopic(instanceId string)

* 数据展示主题, 用以传输实例数据展示算子输出的消息
  DISPLAY/{instance_id}/{node_id}
    - instance_id 展示数据所属实例ID
    - node_id 产生展示数据的节点ID

* 中间消息主题, 用以传输规则实例间的中间消息, 从而联通多媒体与时序规则
  MIDMSG/{pipeline_id}/{midmsg_id}/{instance_id}
    - pipeline_id 中间消息所属规则ID
    - instance_id 中间消息所属实例ID
    - node_id 产生中间消息的节点ID

* 设备消息主题, 用以传输设备上报的数据以及用户下发的指令等
  DATA/{product_id}/{device_id}/{opt_type}/{data_id}
    - product_id 设备所属的产品ID
    - device_id 发出或接收消息的设备的ID
    - opt_type 消息类型:read,write,event,request,response
    - data_id 具体标识一个特定的产品功能

