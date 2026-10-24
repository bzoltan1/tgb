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
	"crypto/subtle"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigPath = "/etc/tgb.yaml"
	maxBodyBytes      = 1 << 20 // 1 MiB
)

type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

type Configuration struct {
	Telegram Telegram `yaml:"telegram"`
	Server   Server   `yaml:"server"`
}

type Telegram struct {
	ChatID int64  `yaml:"chatid"`
	Token  string `yaml:"token"`
}

type Server struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	AuthToken string `yaml:"auth_token"`
}

type handler struct {
	authToken string
	send      func(text string) error
}

func loadConfiguration(path string) (*Configuration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var c Configuration
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if c.Telegram.Token == "" {
		return nil, errors.New("telegram.token is empty")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return nil, fmt.Errorf("server.port out of range: %d", c.Server.Port)
	}
	if c.Server.Host == "" {
		c.Server.Host = "127.0.0.1"
	}
	return &c, nil
}

func configPath() string {
	if p := os.Getenv("TGB_CONFIG"); p != "" {
		return p
	}
	return defaultConfigPath
}

func (h *handler) authorized(r *http.Request) bool {
	if h.authToken == "" {
		return true
	}
	token := r.Header.Get("X-Auth-Token")
	if token == "" {
		if a := r.Header.Get("Authorization"); strings.HasPrefix(a, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(a, "Bearer "))
		}
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(h.authToken)) == 1
}

func (h *handler) sendMessage(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "request body too large or unreadable", http.StatusRequestEntityTooLarge)
		return
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if message.Sender == "" || message.Content == "" {
		http.Error(w, "sender and content are required", http.StatusBadRequest)
		return
	}

	if err := h.send(fmt.Sprintf("%s: %s", message.Sender, message.Content)); err != nil {
		log.Printf("telegram send failed: %v", err)
		http.Error(w, "telegram send failed", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(message)
}

func newRouter(h *handler) http.Handler {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/send", h.sendMessage).Methods(http.MethodPost)
	return myRouter
}

func main() {
	path := flag.String("config", configPath(),
		"path to the YAML config file (default $TGB_CONFIG or "+defaultConfigPath+")")
	flag.Parse()

	c, err := loadConfiguration(*path)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(c.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}

	if c.Server.AuthToken == "" {
		log.Printf("WARNING: server.auth_token is empty; /send is unauthenticated")
	}

	h := &handler{
		authToken: c.Server.AuthToken,
		send: func(text string) error {
			_, err := bot.Send(tgbotapi.NewMessage(c.Telegram.ChatID, text))
			return err
		},
	}

	srv := &http.Server{
		Addr:              net.JoinHostPort(c.Server.Host, strconv.Itoa(c.Server.Port)),
		Handler:           newRouter(h),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Printf("listening on %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}

// curl --request POST   --data '{"sender":"curl","content":"Article Content"}'   http://localhost:10000/send
