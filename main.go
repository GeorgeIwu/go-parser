package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// JSON-RPC response structure
type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
	ID     int             `json:"id"`
}

// Block represents a simplified Ethereum block.
type Block struct {
	Hash         string        `json:"hash"`
	Transactions []Transaction `json:"transactions"`
}

// Transaction represents a simplified Ethereum transaction.
type Transaction struct {
	Hash        string `json:"hash"`
	BlockNumber string `json:"blockNumber"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
}

// MemoryStorage represents an in-memory data storage.
type MemoryStorage struct {
	subscribers map[string]bool // Map from address to subscribers
}

// NewMemoryStorage initializes a new MemoryStorage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		subscribers: make(map[string]bool),
	}
}

func (memory *MemoryStorage) GetSubscribers() (map[string]bool, error) {
	return memory.subscribers, nil
}

func (memory *MemoryStorage) SetSubscriber(address string) error {
	memory.subscribers[address] = true
	return nil
}

func (memory *MemoryStorage) IsSubscriber(address string) bool {
	value, ok := memory.subscribers[address]
	if !ok {
		return false
	}
	return value
}

// Parser defines the interface for interacting with Ethereum blockchain.
type Parser interface {
	GetCurrentBlock() (uint64, error)
	GetTransactions(address string) ([]Transaction, error)
	SubscribeAddress(address string) bool
}

// EthereumParser implements the Parser interface for Ethereum blockchain.
type EthereumParser struct {
	Endpoint string
	memoryDB *MemoryStorage
}

// NewEthereumParser initializes a new EthereumParser instance.
func NewEthereumParser(endpoint string, memoryDB *MemoryStorage) *EthereumParser {
	return &EthereumParser{
		Endpoint: endpoint,
		memoryDB: memoryDB,
	}
}

// GetCurrentBlock gets the current block number from the Ethereum node.
func (parser *EthereumParser) GetCurrentBlock() (uint64, error) {
	var blockNumberHex string
	err := parser.callRPCMethod("eth_blockNumber", nil, &blockNumberHex)
	if err != nil {
		return 0, err
	}

	blockNumber, err := ParseHexUint64(blockNumberHex)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

// GetTransactions queries transactions for an address.
func (parser *EthereumParser) GetTransactions(address string) ([]Transaction, error) {
	var transactions []Transaction
	var block Block
	if !parser.memoryDB.IsSubscriber(address) {
		return nil, fmt.Errorf("Address: %v is not subscribed", address)
	}
	blockNumber, err := parser.GetCurrentBlock()
	if err != nil {
		return nil, err
	}
	err = parser.callRPCMethod("eth_getBlockByNumber", ParseToAnySlice(fmt.Sprintf("0x%x", blockNumber), true), &block)
	if err != nil {
		return nil, err
	}

	for _, transaction := range block.Transactions {
		if transaction.From == address || transaction.To == address {
			transactions = append(transactions, transaction)
		}
	}

	return transactions, nil
}

// SubscribeAddress subscribes to an Ethereum address.
func (parser *EthereumParser) SubscribeAddress(address string) bool {
	if err := parser.memoryDB.SetSubscriber(address); err != nil {
		return false
	}
	return true
}

// callRPCMethod sends a JSON-RPC request to the Ethereum node.
func (parser *EthereumParser) callRPCMethod(method string, params []interface{}, result interface{}) error {
	var response RPCResponse
	requestBody := fmt.Sprintf(`{
		"jsonrpc": "2.0",
		"method": "%s",
		"params": %s,
		"id": 1
	}`, method, toJSON(params))

	resp, err := http.Post(parser.Endpoint, "application/json", strings.NewReader(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		fmt.Printf("Failed to decode JSON-RPC response: %v\n", err)
		return err
	}

	// Check for errors in response
	if response.Error != nil {
		return fmt.Errorf("JSON-RPC error: %v", response.Error)
	}

	// parse to result
	err = json.Unmarshal(response.Result, &result)
	if err != nil {
		return fmt.Errorf("failed to parse response details: %v", err)
	}

	return nil
}

// toJSON converts parameters to JSON string.
func toJSON(params []interface{}) string {
	if len(params) == 0 {
		return "[]"
	}

	var builder strings.Builder
	builder.WriteByte('[')
	for i, param := range params {
		jsonParam, _ := json.Marshal(param)
		builder.Write(jsonParam)
		if i < len(params)-1 {
			builder.WriteByte(',')
		}
	}
	builder.WriteByte(']')
	return builder.String()
}

// ParseHexUint64 parses a hex-encoded string into a uint64.
func ParseHexUint64(hexStr string) (uint64, error) {
	return strconv.ParseUint(hexStr[2:], 16, 64)
}

// ParseToAnySlice parses any argument string into an interface{}.
func ParseToAnySlice(params ...interface{}) []interface{} {
	var allParams []interface{}

	// Convert each string element to interface and append to allParams
	for _, param := range params {
		allParams = append(allParams, param)
	}

	return allParams
}

func processCommands(cmdCh <-chan string) {
	// Ethereum node JSON-RPC endpoint (replace with your own endpoint)
	endpoint := "https://cloudflare-eth.com"

	memoryDB := NewMemoryStorage()

	// Create EthereumParser instance
	parser := NewEthereumParser(endpoint, memoryDB)

	var args []string
	for {
		select {
		case cmd := <-cmdCh:
			args = strings.Fields(cmd)
			if len(args) < 1 {
				fmt.Println("\nYou need to define an action (getCurrentBlock, getTransaction, SubscribeAddress)")
				continue
			}
			action := args[0]

			address := ""
			if len(args) > 1 {
				address = args[1]
			}

			// Example usage
			switch action {
			case "getCurrentBlock":
				blockNumber, err := parser.GetCurrentBlock()
				if err != nil {
					fmt.Printf("\nError getting current block: %v", err)
					continue
				}
				fmt.Println(blockNumber)
				continue
			case "getTransaction":
				if address == "" {
					fmt.Println("\nYou need to define an address")
					continue
				}
				transactions, err := parser.GetTransactions(address)
				if err != nil {
					fmt.Printf("\nError getting transactions: %v", err)
					continue
				}
				fmt.Println(transactions)
				continue
			case "subscribeAddress":
				if address == "" {
					fmt.Println("\nYou need to define an address")
					continue
				}
				isSubscribed := parser.SubscribeAddress(address)
				if !isSubscribed {
					fmt.Printf("\nError subscribing to address: %v", address)
					continue
				}
				fmt.Println(isSubscribed)
				continue
			default:
				fmt.Printf("Invalid action: %v. please pick valid action (getCurrentBlock, getTransaction, subscribeAddress)", action)
				continue
			}
		}
	}

}

func main() {
	// Create a channel to receive commands
	cmdCh := make(chan string)

	// Start a goroutine to continuously process commands
	go processCommands(cmdCh)

	// Main loop to read user input and send commands to the channel
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter command (e.g: getCurrentBlock): ")
		if !scanner.Scan() {
			break
		}

		command := scanner.Text()
		cmdCh <- command // Send the command to the channel
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading standard input: %v\n", err)
	}
}

// The MemoryStorage struct provides a basic in-memory storage for suubscribers. You can extend this by implementing persistent storage (e.g., using a database) by modifying the MemoryStorage methods.
// Error handling is simplified and no tests added for demonstration purposes. In production code, should handle errors more robustly and wrrite tests for all edge cases.
