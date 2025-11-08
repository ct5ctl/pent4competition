# **腾讯云黑客松-智能渗透****挑战赛API文档**

## **概述**

本文档提供智能渗透挑战赛的完整API接口说明，参赛者需要使用这些API来获取赛题信息、查看提示以及提交flag。

##  

## **认证方式**

所有API请求都需要在请求头中包含认证信息：



```
Authorization: Bearer {API_TOKEN}
```





请妥善保管您的API_TOKEN，不要泄露给他人。

##  

## **基础信息**

​                ● **Base URL**: http://x.x.x.x:8000（待公布）

​                ● **Content-Type**: application/json

​                ● **Accept**: application/json

##  

## **API接口列表**

### **1. 获取当前阶段赛题列表**

获取当前阶段的所有赛题信息，包括赛题代码、难度、分值、目标服务器信息以及答题状态。智能体可根据current_stage字段判断比赛是否已开始。

**接口信息**

​                ● **方法**: GET

​                ● **路径**: /api/v1/challenges

​                ● **描述**: 用于获取当前所处阶段（调试阶段/答题阶段），以及当前阶段的赛题列表，包含该赛题的ip端口和答题情况（调试阶段会返回分值为0的demo赛题）

**请求示例**



```
curl -X 'GET' 'http://x.x.x.x:8000/api/v1/challenges' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {API_TOKEN}'
```





**响应示例**

**成功响应（200）**



```
{
  "current_stage": "debug",
  "challenges": [
    {
      "challenge_code": "debugdemo1",
      "difficulty": "medium",
      "points": 300,
      "hint_viewed": false,
      "solved": false,
      "target_info": {
        "ip": "10.0.0.200",
        "port": [
          8080
        ]
      }
    },
    {
      "challenge_code": "debugdemo2",
      "difficulty": "hard",
      "points": 500,
      "hint_viewed": true,
      "solved": false,
      "target_info": {
        "ip": "10.0.0.201",
        "port": [
          80
        ]
      }
    }
  ]
}
```





**字段说明**

| **字段**                       | **类型** | **说明**                                                     |
| ------------------------------ | -------- | ------------------------------------------------------------ |
| current_stage                  | string   | 当前所处的阶段（debug：调试阶段，competition：正式答题阶段） |
| challenges[x].challenge_code   | string   | 赛题唯一标识码                                               |
| challenges[x].difficulty       | string   | 难度等级（easy/medium/hard）                                 |
| challenges[x].points           | integer  | 该题目的基础分值                                             |
| challenges[x].hint_viewed      | boolean  | 是否已查看过提示                                             |
| challenges[x].solved           | boolean  | 是否已成功解答                                               |
| challenges[x].target_info.ip   | string   | 目标服务器IP地址                                             |
| challenges[x].target_info.port | array    | 目标服务器端口列表                                           |

### **2. 查看提示**

获取指定赛题的提示信息。查看提示后，成功解题时会在得分时扣除一定惩罚积分。每题仅有一个提示，多次调用不会重复扣分。

**接口信息**

​                ● **方法**: GET

​                ● **路径**: /api/v1/hint/{challenge_code}

​                ● **描述**: 查看题目提示信息，查看提示后，成功解题时会在得分时扣除一定惩罚积分

**路径参数**

| **参数**       | **类型** | **必填** | **说明**       |
| -------------- | -------- | -------- | -------------- |
| challenge_code | string   | 是       | 赛题唯一标识码 |

**请求示例**



```
curl -X 'GET' 'http://10.0.0.6:8000/api/v1/hint/{challenge_code}' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {API_TOKEN}'
```





**响应示例**

**首次查看（200）**



```
{
  "hint_content": "Every time you see a value that looks like an ID, you should try to fuzz it with similar values!",
  "penalty_points": 30,
  "first_use": true
}
```





**非首次查看（200）**



```
{
  "hint_content": "Every time you see a value that looks like an ID, you should try to fuzz it with similar values!",
  "penalty_points": 30,
  "first_use": false
}
```





**错误响应**

**题目不存在（500）**



```
{
  "detail": "题目不存在"
}
```





**题目已经完成解答（500）**



```
{
  "detail": "该题目已成功解答"
}
```





**字段说明**

| **字段**       | **类型** | **说明**           |
| -------------- | -------- | ------------------ |
| hint_content   | string   | 提示内容           |
| penalty_points | integer  | 查看提示的惩罚积分 |
| first_use      | boolean  | 是否为首次查看     |

### **3. 提交答案**

提交赛题答案并获得积分。

**接口信息**

​                ● **方法**: POST

​                ● **路径**: /api/v1/answer

​                ● **描述**: 提交答案、获得积分（针对一个题目，最多仅支持答题100次）。答案为flag{}或FLAG{}包裹的字符串，“flag”大小写均可。

**请求体参数**

| **参数**       | **类型** | **必填** | **说明**                   |
| -------------- | -------- | -------- | -------------------------- |
| challenge_code | string   | 是       | 赛题唯一标识码             |
| answer         | string   | 是       | 答案内容（通常为flag格式） |

**请求示例**



```
curl -X 'POST' 'http://10.0.0.6:8000/api/v1/answer' \
  -H 'accept: application/json' \
  -H 'Authorization: Bearer {API_TOKEN}' \
  -H 'Content-Type: application/json' \
  -d '{
    "challenge_code": "debugdemo1",
    "answer": "flag{e82fba66-3230-4818-8960-800ab47b0ba3}"
  }'
```





**响应示例**

**回答正确（200）**



```
{
  "correct": true,
  "earned_points": 270,
  "is_solved": false
}
```





**回答正确重复提交（200）**



```
{
  "correct": true,
  "earned_points": 270,
  "is_solved": true
}
```





**回答错误（200）**



```
{
  "correct": false,
  "earned_points": 0,
  "is_solved": false
}
```





**错误响应**

**题目不存在（500）**



```
{
  "detail": "题目不存在"
}
```





**字段说明**

| **字段**      | **类型** | **说明**                         |
| ------------- | -------- | -------------------------------- |
| correct       | boolean  | 答案是否正确                     |
| earned_points | integer  | 本次获得的积分（已扣除惩罚积分） |
| is_solved     | boolean  | 该题目之前是否已被解答           |

##  

## **API异常响应说明**

当API调用出现错误时，系统会返回相应的错误码和错误信息，请根据以下说明进行处理。

### **错误码列表**

#### **401 - 认证失败**

**描述**: 请求头Authorization中设置的API_TOKEN或格式有误

**响应示例**



```
{
  "detail": "认证错误"
}
```





**解决方案**

​                ● 检查Authorization请求头格式是否正确

​                ● 确认API_TOKEN是否正确

#### **422 - 参数有误**

**描述**: 必填参数缺失或格式有误

**响应示例**



```
{
  "detail": "参数校验失败"
}
```





**解决方案**

​                ● 根据接口文档检查参数是否有误

#### **429 - 请求过于频繁**

**描述**: 接口请求频率超限，所有接口默认频率限制在1次/秒，答题接口针对同一题目，请求次数限制为100次，超限会响应429。

**响应示例**



```
{
  "detail": "请求过于频繁"
}
```





**解决方案**

​                ● 检查请求逻辑，避免重复请求

​                ● 增加一定请求的时间间隔

#### **500 - 接口业务异常**

**描述**: 根据具体接口响应的详细描述进行处理

**响应示例**



```
{
  "detail": "题目不存在"
}
```





**解决方案**

​                ● 检查参数是否有误

​                ● 检查调用场景是否符合接口定义

##  

## **注意事项**

​            1.     **响应码说明**: 仅在响应码为200、500时，请求格式合法，若出现其他响应码，可检查参数字段

​            2.     **请求频率**: 请根据接口请求频率限制进行适当的请求间隔，若出现429响应码，请检查请求逻辑

​            3.     **积分计算**: 查看提示后成功解题会扣除惩罚积分，最终得分 = 基础分值 - 惩罚积分

​            4.     **重复提交**: 同一题目重复提交正确答案不会重复获得积分

​            5.     **API_TOKEN安全**: 请妥善保管您的API_TOKEN，不要泄露给他人

##  

## **使用流程**

​            1.     **获取赛题列表**: 调用 /api/v1/challenges 获取当前可用的赛题

​            2.     **分析赛题**: 根据返回的目标服务器信息（IP和端口）进行安全测试

​            3.     **查看提示（可选）**: 如需帮助，可调用 /api/v1/hint/{challenge_code} 查看提示

​            4.     **提交答案**: 找到flag后，调用 /api/v1/answer 提交答案获取积分

##  

## **技术支持**

如遇到技术问题，请及时联系技术支持团队。