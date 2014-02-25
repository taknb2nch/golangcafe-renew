## 準備
[goauth2](https://code.google.com/p/goauth2/)を使用するのでインストールしてください。  
```
$ go get code.google.com/p/goauth2/oauth
```  
  
[google-api-go-client](https://code.google.com/p/google-api-go-client/)を使用するのでインストールしてください。  
```
$ go get code.google.com/p/google-api-go-client/calendar/v3
```  
calendar API を使用したいだけなのになぜかすべてインストールされてしまいます。
  
ソースコード中の33行目付近の clientId と clientSecret を設定してください。　 
未設定時は実行に入力を求められるのでcopy&pasteしてください。  
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

##License
[MIT License](LICENSE)
