package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"google.golang.org/genai"
)

// 預先定義好 HTML 訊息片段的模板
var msgTemplate = template.Must(template.New("messages").Parse(`
	<div class="message user-message">{{.UserMsg}}</div>
	<div class="message bot-message">{{.BotReply}}</div>
`))

type ReplyData struct {
	UserMsg  string
	BotReply string
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatalf("無法初始化 Gemini 用戶端: %v", err)
	}

	// 1. 路由：首頁，直接奉送 index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// 2. 路由：HTMX 呼叫的 API
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// HTMX 預設是用傳統表單格式 (Form Value) 傳送資料，不是 JSON 了
		userMessage := r.FormValue("message")
		if userMessage == "" {
			return
		}

		// 呼叫 Gemini
		resp, err := client.Models.GenerateContent(r.Context(), "gemini-2.5-flash", genai.Text(userMessage), nil)
		botReply := ""
		if err != nil {
			botReply = fmt.Sprintf("錯誤: 暫時無法獲取回應 (%v)", err)
		} else {
			botReply = resp.Text()
		}

		// 直接渲染 HTML 片段回傳給 HTMX
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := ReplyData{
			UserMsg:  userMessage,
			BotReply: botReply,
		}
		msgTemplate.Execute(w, data)
	})

	fmt.Printf("HTMX 伺服器運行於 port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
