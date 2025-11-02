package sse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Event struct {
	Name  string
	ID    string
	Retry int
	Data  any
}

func (e Event) WriteTo(w io.Writer) error {
	bw := bufio.NewWriter(w)
	if e.Name != "" {
		if _, err := fmt.Fprintf(bw, "event: %s\n", e.Name); err != nil {
			return err
		}
	}
	if e.ID != "" {
		if _, err := fmt.Fprintf(bw, "id: %s\n", e.ID); err != nil {
			return err
		}
	}
	if e.Retry > 0 {
		if _, err := fmt.Fprintf(bw, "retry: %d\n", e.Retry); err != nil {
			return err
		}
	}
	if e.Data != nil {
		b, err := json.Marshal(e.Data)
		if err != nil {
			return err
		}
		// Split by newlines just in case; SSE requires each line prefixed by "data: "
		lines := splitLines(string(b))
		for _, line := range lines {
			if _, err := fmt.Fprintf(bw, "data: %s\n", line); err != nil {
				return err
			}
		}
	}
	// End of message
	if _, err := fmt.Fprint(bw, "\n"); err != nil {
		return err
	}
	return bw.Flush()
}

// Heartbeat emits a tiny SSE keep-alive record to help keep long-lived HTTP
// HTTP層（L7）。TCP の Keep-Aliveではアイドル判定が期待通りに動かずにない可能性があるため、定期的にHTTPボディへ最少バイトを送る
func Heartbeat(w io.Writer, msg string) error {
	bw := bufio.NewWriter(w)
	if msg == "" {
		msg = "keep-alive"
	}
	// 先頭が : の行は SSEの“コメント”で、ブラウザの EventSource ではイベントとしては発火しない
	if _, err := fmt.Fprintf(bw, ": %s %d\n\n", msg, time.Now().Unix()); err != nil {
		return err
	}
	return bw.Flush()
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}
