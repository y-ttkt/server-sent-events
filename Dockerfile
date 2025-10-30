FROM golang:1.23.5
ENV TZ=Asia/Tokyo

WORKDIR /app

# 依存が追加された後にgo.sumもCOPY対象とする
COPY go.mod ./
RUN go mod download

EXPOSE 8080