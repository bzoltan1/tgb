package main

import (
    "fmt"
    "log"
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
    "io/ioutil"
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "gopkg.in/yaml.v2"

)

type Message struct {
    Sender   string `json:"Sender"`
    Content string `json:"content"`
}

type Configuration struct {
    Telegram Telegram
    Server  Server
}

type Telegram struct {
    ChatID  int64
    Token   string
}

type Server struct {
    Host   string
    Port   int
}

var bot *tgbotapi.BotAPI
var err error
var c Configuration

func (c *Configuration) getConfiguration() *Configuration {
    yamlFile, err := ioutil.ReadFile("tgb.yaml")
    if err != nil {
        log.Printf("yamlFile.Get err   #%v ", err)
    }
    err = yaml.Unmarshal(yamlFile, c)
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }
    return c
}


func sendMessage(w http.ResponseWriter, r *http.Request) {
    reqBody, _ := ioutil.ReadAll(r.Body)
    var message Message
    json.Unmarshal(reqBody, &message)
    json.NewEncoder(w).Encode(message)
    msg := tgbotapi.NewMessage(c.Telegram.ChatID, fmt.Sprintf("%s: %s",message.Sender, message.Content))
    bot.Send(msg)
}


func handleRequests() {
    myRouter := mux.NewRouter().StrictSlash(true)
    myRouter.HandleFunc("/send", sendMessage).Methods("POST")
    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.Server.Port), myRouter))
}

func main() {

    c.getConfiguration()
    bot,err = tgbotapi.NewBotAPI(c.Telegram.Token)
    if err != nil {
        log.Panic(err)
    }
    handleRequests()
}
// curl --request POST   --data '{"sender":"curl","content":"Article Content"}'   http://localhost:10000/send
