# tgb
Telegram bridge
# A service what acts as a bridge between anything and Telegram
I just had an idea that I would like to see some logs and events not just dumped on the /var/log space but to send to my Telegram account.

Initially I was planning to use it as alert in monit services.
But anything can send Telegram messages by a simple Rest API call to the server
```
$ curl --request POST   --data '{"sender":"curl","content":"Article Content"}'   http://localhost:10000/send
```
