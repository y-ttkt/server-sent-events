# server-sent-events

サーバー送信イベント（SSE: Server-Sent Events）を Go（net/http）で実装した最小構成のサンプルです。
1本のHTTPレスポンスを開きっぱなしにして、サーバからクライアント（ブラウザ）へテキストを逐次配信します。

## エンドポイント
- GET /stream… SSE ストリーム（text/event-stream）

- GET /healthz… ヘルスチェック

- GET /… 動作確認用のテストページ

## クイックスタート
```
# 1. 取得
git clone git@github.com:y-ttkt/server-sent-events.git
cd server-sent-events
```
```
# 2. 初期セットアップ
docker compose build
```
```
# 3. 動作確認
docker compose up -d
go run cmd/server/main.go
curl http://localhost:8080/healthz
```

## ディレクトリ構成
```
.
├─ cmd/server/main.go      # ルータ/SSE/心拍/Graceful shutdown
├─ internal/sse/sse.go     # Event整形/Heartbeat/WriteTo
├─ web/index.html          # 動作確認ページ
├─ nginx/nginx.conf        # リバースプロキシの設定
├─ compose.yml             # 簡易ローカル用
├─ Dockerfile              # 簡易ローカル用
├─ Makefile                # run/build/tidy 等
└─ README.md
```