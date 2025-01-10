# cbum

To run the project clone the project and do `docker compose up`

To test it use following data in postman - 

`User1` - 
```json
{
    "MessageType" : "loginmsg",
	"Data" :  "",
	"Reciever" : "",
	"Sender" : "tan"
}
```

`User2` - 
```json
{
    "MessageType" : "loginmsg",
	"Data" :  "",
	"Reciever" : "",
	"Sender" : "yan"
}
```

`user2` - 
```json
{
    "MessageType" : "chatmsg",
	"Data" :  "hiii",
	"Reciever" : "tan",
	"Sender" : "yan"
}
```