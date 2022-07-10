## 1. lists

- 方法描述：查询该集群下的实例列表
- URL地址：/api/instance/lists
- 请求方式：get
- 请求参数：

| 字段    | 类型   | 说明     | 是否必传 |
| :------ | :----- | :------- | :------- |
| group_name | string | 组名称 | Y        |



- 返回参数

| 字段                    | 类型      | 说明     | json key    |
| :---------------------- | :-------- | :------- | :---------- |
| Errcode               | RetHeader | 返回头   | errcode  |
| Errmsg               | RetHeader | 返回头   | errmsg  |
| Result                    | string[]  | 返回值   | result    |


## 2.add

- 方法描述：add实例
- URL地址：/api/instance/add
- 请求方式：post
- 请求参数


- 请求body

| 字段      | 类型      | 说明                           | 是否必传 |
| :-------- | :-------- | :--------------------------    | :------- |
| group_name | form      | 组名 | Y        |
| ip | form      | ip地址 | Y        |
| port | form      | port | Y        |
| addr | form      | ip:port | Y        |
| probe_type | form      | 探活类型 tcp、http两种 | Y        |
| probe_url | form      | probe_type=http时 要填的探活接口 | N        |


- 返回参数

| 字段       | 类型   | 说明     | json key    |
| :--------- | :----- | :------- | :---------- |
| Errcode               | RetHeader | 返回头   | errcode  |
| Errmsg               | RetHeader | 返回头   | errmsg  |



## 3.remove

- 方法描述：删除实例
- URL地址：/api/instance/remove
- 请求方式：delete
- 请求参数

| 字段    | 类型   | 说明                  | 是否必传 |
| :------ | :----- | :-------------------- | :------- |
| group_name | form      | 组名 | Y        |
| ip | form      | ip地址 | Y        |
| port | form      | port | Y        |
| addr | form      | ip:port | Y        |
| probe_type | form      | 探活类型 tcp、http两种 | Y        |
| probe_url | form      | probe_type=http时 要填的探活接口 | N        |



- 返回参数

| 字段       | 类型   | 说明     | json key    |
| :--------- | :----- | :------- | :---------- |
| Errcode               | RetHeader | 返回头   | errcode  |
| Errmsg               | RetHeader | 返回头   | errmsg  |

