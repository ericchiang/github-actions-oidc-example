package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

const actionsIDTokenRequestURLEnv = "ACTIONS_ID_TOKEN_REQUEST_URL"
const actionsIDTokenRequestTokenEnv = "ACTIONS_ID_TOKEN_REQUEST_TOKEN"

func run(ctx context.Context) error {
	url := os.Getenv(actionsIDTokenRequestURLEnv)
	token := os.Getenv(actionsIDTokenRequestTokenEnv)
	fmt.Println(url, token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read all: %w", err)
	}
	fmt.Println(string(body))
	parts := strings.Split(string(body), ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid body: %s", string(body))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}
	buf := bytes.NewBuffer(nil)
	if err := json.Indent(buf, payload, "", "  "); err != nil {
		return fmt.Errorf("indent: %w", err)
	}
	fmt.Println(buf.String())
	return nil
}
