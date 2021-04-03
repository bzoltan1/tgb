// Copyright 2021 Zoltán Balogh.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the  General Public License as published by
// the Free Software Foundation; version 2.0
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// Author: Zoltán Balogh <zoltan@bakter.hu>

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
