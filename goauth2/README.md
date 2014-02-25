## 準備
[goauth2](https://code.google.com/p/goauth2/)を使用するのでインストールしてください。  
```
$ go get code.google.com/p/goauth2/oauth
```  

ソースコード中の33行目付近の clientId と clientSecret を設定してください。　
```
var (
    ...略

    // clientID、secret、はDevelopers ConsoleのCredentialsからコピー＆ペーストして下さい。
    clientId     = ""
    clientSecret = ""

    ...略
)
```

##実行
```
$ go run getcalendarlist.go
```  