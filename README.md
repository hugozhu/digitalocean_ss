# 介绍

通过脚本从[DigitalOcean](https://m.do.co/c/e85760c85e44)的snapshot生成一个新的虚拟主机（snapshot可以预先设置好VPN或SS服务），并绑定到对应的dnspod记录。

## 设置
当前目录下编辑`config.json`配置文件，参考内容如下：
```
{
    "token":"<your_digital_ocean_access_token>",
    "domain":"<your_domain>",
    "dnspod": {
    	"login_email":"<dnspod_login_email>",
    	"login_password":"<dnspod_login_password>",
    	"domain_id":123456,
    	"record_id":456789,
    	"sub_domain":"sf"
    }
}
```

##  生成并启动服务器

```
go run src/main.go -create true
```

##  关闭并销毁服务器
销毁后将不再计费

```
go run src/main.go -destroy true
```