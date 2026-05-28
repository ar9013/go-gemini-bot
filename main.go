package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"google.golang.org/genai"
)

// 定義請求與回應的結構
type ChatRequest struct {
	Message string `json:"message"`
}

type ChatResponse struct {
	Reply string `json:"reply"`
}

func main() {
	// 1. 從環境變數獲取 Port，Railway 會自動注入這個變數
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. 初始化 Gemini 用戶端
	// SDK 會自動從環境變數 GEMINI_API_KEY 中讀取金鑰
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatalf("無法初始化 Gemini 用戶端: %v", err)
	}

	// 3. 設定路由
	http.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		// 僅允許 POST 請求
		if r.Method != http.MethodPost {
			http.Error(w, "只支援 POST 請求", http.StatusMethodNotAllowed)
			return
		}

		// 解析前端傳來的 JSON
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "無效的 JSON 格式", http.StatusBadRequest)
			return
		}

		// 呼叫 Gemini 模型 (使用推薦的 gemini-2.5-flash)
		resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(req.Message), nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("AI 產生回應失敗: %v", err), http.StatusInternalServerError)
			return
		}

		// ✨ 修正這裡：正確提取並轉換文字回應
		var replyText string
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			for _, part := range resp.Candidates[0].Content.Parts {
				// 新版 SDK 提供 Text 欄位，或者我們可以使用型態斷言來安全取得字串
				if textPart, ok := part.(genai.Text); ok {
					replyText += string(textPart)
				} else {
					// 備用方案：如果不是 Text 型態，再嘗試用 fmt.Sprint 轉換
					replyText += fmt.Sprint(part)
				}
			}
		}

		// 回傳 JSON 給前端
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChatResponse{Reply: replyText})
	})

	// 監聽並啟動服務
	fmt.Printf("伺服器正在運行於 port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
