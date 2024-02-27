# ChatGPT-API-SERVER

使用 chat 接口转换对外提供API服务，支持多账号轮循，支持并发控制，支持多种模型。

 

[![Open in GitHub Codespaces](https://github.com/codespaces/badge.svg)](https://github.com/codespaces/new?hide_repo_select=true&ref=master&repo=747643293)

## gf工具安装
```shell
go install github.com/gogf/gf/cmd/gf/v2@latest 
```
- 提交docker 账号是用 hts999 
- 开发机上 都先要 `docker logout` 再 `docker login`
- 这边代码布置的前端不能用 可能要到作者的前端里面copy出来

## 环境变量

```yaml
# 接入网关地址
CHATPROXY: "https://demo.xyhelper.cn"
# 接入网关的authkey
AUTHKEY: "xyhelper"
# 是否限制每个用户只能同时发起一个会话
USERTOKENLOCK: true
# 3.5模型是否免费
ISFREE: true
```
## 调试
先在host里面增加下myql redis
```shell
gf run main.go
```
