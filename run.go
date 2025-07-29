package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	tokenURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	if tokenURL == "" {
		return fmt.Errorf("ACTIONS_ID_TOKEN_REQUEST_URL is not set, need to set 'id-token: write' permission?")
	}
	token := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	if token == "" {
		return fmt.Errorf("ACTIONS_ID_TOKEN_REQUEST_TOKEN is not set, need to set 'id-token: write' permission?")
	}
	u, err := url.Parse(tokenURL)
	if err != nil {
		return fmt.Errorf("parsing token request url: %w", err)
	}

	// Set the audience.
	q := u.Query()
	q.Set("audience", "oblique.security")
	u.RawQuery = q.Encode()

	// Create the request and set the authorization header.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed %s: %s", resp.Status, string(body))
	}
	var respBody struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(body, &respBody); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	rawIDToken := respBody.Value

	provider, err := oidc.NewProvider(ctx, "https://token.actions.githubusercontent.com")
	if err != nil {
		return fmt.Errorf("new provider: %w", err)
	}
	config := &oidc.Config{
		ClientID: "oblique.security",
	}
	verifier := provider.Verifier(config)

	// Verify the ID Token.
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return fmt.Errorf("verifying token: %w", err)
	}
	var claims struct {
		Actor       string `json:"actor"`
		Environment string `json:"environment"`
		Ref         string `json:"ref"`
		Repository  string `json:"repository"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return fmt.Errorf("claims: %w", err)
	}
	fmt.Println("Actor:", claims.Actor)
	fmt.Println("Subject:", idToken.Subject)
	fmt.Println("Ref:", claims.Ref)
	fmt.Println("Repository:", claims.Repository)
	return nil
}
