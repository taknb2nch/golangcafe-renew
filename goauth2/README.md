## 準備
###パッケージのインストール
[goauth2](https://code.google.com/p/goauth2/)を使用するのでインストールしてください。  
```
$ go get code.google.com/p/goauth2/oauth
```  
  
[google-api-go-client](https://code.google.com/p/google-api-go-client/)を使用するのでインストールしてください。  
```
$ go get code.google.com/p/google-api-go-client/calendar/v3
```  
calendar API を使用したいだけなのになぜかすべてインストールされてしまいます。
  
###ソースコードの編集
ソースコード中の33行目付近の clientId と clientSecret を設定してください。（任意）   
未設定時は実行に入力を求められるので必須ではありません。  
```
var (
    ...略

    // clientID、secret、はDevelopers ConsoleのCredentialsからコピー＆ペーストして下さい。
    clientId     = ""
    clientSecret = ""

    ...略
)
```  

###APIの有効化
[Google Developers Console](https://cloud.google.com/console/project)でCalendar APIを有効にしてください。  
設定方法は http://d.hatena.ne.jp/taknb2nch/20140225/1393314429 を参考にしてください。  

##実行
```
$ go run getcalendarlist.go
```  
