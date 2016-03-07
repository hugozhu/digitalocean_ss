# 设置

## config.json

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

```
go run src/main.go -destroy true
```