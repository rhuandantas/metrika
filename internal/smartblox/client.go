package smartblox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Metrika-Inc/smartblox"
)

// Status models /api/status
type Status struct {
	LastRound int64 `json:"last-round"`
}

// Transaction models the inner transaction.
type Transaction struct {
	Amount     int64  `json:"amount"`
	Sender     int64  `json:"sender"`
	Type       string `json:"type"`
	Receipient int64  `json:"receipient"`
}

// TransactionSig is the outer entry carrying the tx + signature.
type TransactionSig struct {
	Sig string      `json:"sig"`
	Tx  Transaction `json:"tx"`
}

// Block models /api/blocks/{round}
type Block struct {
	Round int64            `json:"round"`
	Txs   []TransactionSig `json:"txs"`
}

// Client defines the interface for a SmartBlox API client.
type Client interface {
	// GetStatus fetches the current status from the SmartBlox API.
	// If mockSmartBloxAPI is true, it simulates the API response.
	// Params:
	//   - ctx: Context for request cancellation and timeouts.
	// Returns:
	//   - Status: The fetched status data.
	//   - error: An error if the operation fails.
	GetStatus(ctx context.Context) (Status, error)
	// GetBlock fetches a specific block by round from the SmartBlox API.
	// If mockSmartBloxAPI is true, it simulates the API response.
	// Params:
	//   - ctx: Context for request cancellation and timeouts.
	//   - round: The block round number to fetch.
	// Returns:
	//   - Block: The fetched block data.
	//   - error: An error if the operation fails.
	GetBlock(ctx context.Context, round int64) (Block, error)
}

type httpClient struct {
	base             string
	c                *http.Client
	mockSmartBloxAPI bool
}

func NewHTTPClient(base string, timeout time.Duration, mockSmartBloxAPI bool) Client {
	return &httpClient{
		base:             base,
		c:                &http.Client{Timeout: timeout},
		mockSmartBloxAPI: mockSmartBloxAPI,
	}
}

// GetStatus fetches the current status from the SmartBlox API.
func (h *httpClient) GetStatus(ctx context.Context) (Status, error) {
	if h.mockSmartBloxAPI {
		return mockGetStatus()
	}

	url := fmt.Sprintf("%s/api/status", h.base)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := h.c.Do(req)
	if err != nil {
		return Status{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Status{}, fmt.Errorf("status code %d", resp.StatusCode)
	}
	var s Status
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return Status{}, err
	}
	return s, nil
}

// mockGetStatus simulates fetching status from the SmartBlox API.
func mockGetStatus() (Status, error) {
	status, err := smartblox.GetStatus()
	if err != nil {
		return Status{}, err
	}

	var s Status
	err = json.Unmarshal(status, &s)
	if err != nil {
		return Status{}, err
	}

	return s, nil
}

func (h *httpClient) GetBlock(ctx context.Context, round int64) (Block, error) {
	if h.mockSmartBloxAPI {
		return mockGetBlock(round)
	}

	url := fmt.Sprintf("%s/api/blocks/%d", h.base, round)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := h.c.Do(req)
	if err != nil {
		return Block{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Block{}, fmt.Errorf("status code %d", resp.StatusCode)
	}
	var b Block
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return Block{}, err
	}
	return b, nil
}

// mockGetBlock simulates fetching a block by round from the SmartBlox API.
func mockGetBlock(round int64) (Block, error) {
	block, err := smartblox.GetBlock(round)
	if err != nil {
		return Block{}, err
	}

	var b Block
	err = json.Unmarshal(block, &b)
	if err != nil {
		return Block{}, err
	}

	return b, nil
}
