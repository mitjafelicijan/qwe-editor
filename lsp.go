package main

// Basic Language Server Protocol (LSP) client. Communicates with external
// language servers (like gopls or clangd) via JSON-RPC over standard
// input/output.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nsf/termbox-go"
)

// LSPClient manages the lifecycle and communication with an LSP server process.
type LSPClient struct {
	cmd          *exec.Cmd      // The underlying server process.
	stdin        io.WriteCloser // Write messages to the server.
	stdout       io.ReadCloser  // Read messages from the server.
	scanner      *bufio.Scanner
	messageID    int64        // Monotonically increasing ID for requests.
	diagnostics  []Diagnostic // Cached errors/warnings from the server.
	diagMutex    sync.RWMutex // Protects access to diagnostics.
	filename     string       // The file this client is associated with.
	uri          string       // The LSP-compatible URI of the file.
	shutdown     bool         // Flag to indicate the client is closing.
	shutdownOnce sync.Once
	logCallback  func(string, string) // Debug logging.

	responses     map[int64]chan map[string]interface{} // Map of request IDs to response channels.
	responseMutex sync.Mutex
	fileType      *FileType // Associated file type for language ID.
}

// Position in a document (0-based line and character).
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range represents a span of text in a document.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location points to a specific range in a specific file.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// CompletionItem represents a suggestion for completion.
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"`
	Detail        string `json:"detail"`
	Documentation string `json:"documentation"`
	InsertText    string `json:"insertText"`
}

// CompletionList represents a collection of completion items.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// Diagnostic represents an error, warning, or hint from the language server.
type Diagnostic struct {
	Range struct {
		Start struct {
			Line      int `json:"line"`
			Character int `json:"character"`
		} `json:"start"`
		End struct {
			Line      int `json:"line"`
			Character int `json:"character"`
		} `json:"end"`
	} `json:"range"`
	Severity int    `json:"severity"` // 1=Error, 2=Warning, 3=Info, 4=Hint.
	Message  string `json:"message"`
}

// NewLSPClient starts a new LSP server process for the given file type.
func NewLSPClient(filename string, fileContent string, logCallback func(string, string), ft *FileType) (*LSPClient, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	client := &LSPClient{
		filename:    absPath,
		uri:         "file://" + absPath,
		diagnostics: []Diagnostic{},
		logCallback: logCallback,
		responses:   make(map[int64]chan map[string]interface{}),
		fileType:    ft,
	}

	// Launch the language server's executable.
	client.cmd = exec.Command(ft.LSPCommand, ft.LSPCommandArgs...)

	// Suppress the server's own internal log messages (stderr).
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err == nil {
		client.cmd.Stderr = devNull
	}

	client.stdin, err = client.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	client.stdout, err = client.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := client.cmd.Start(); err != nil {
		return nil, err
	}

	// Start a background goroutine to read messages from the server's stdout.
	go client.readMessages()

	// Perform the LSP handshake: Initialize and Notify Open.
	if err := client.initialize(); err != nil {
		client.Shutdown()
		return nil, err
	}

	if err := client.sendDidOpen(fileContent); err != nil {
		client.Shutdown()
		return nil, err
	}

	return client, nil
}

// nextID increments and returns the next request ID.
func (c *LSPClient) nextID() int64 {
	return atomic.AddInt64(&c.messageID, 1)
}

// sendRequest sends a JSON-RPC request and expects a response.
func (c *LSPClient) sendRequest(method string, params interface{}) error {
	id := c.nextID()
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	return c.sendMessage(request)
}

// sendNotification sends a JSON-RPC message without expecting a response.
func (c *LSPClient) sendNotification(method string, params interface{}) error {
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}
	return c.sendMessage(notification)
}

// sendMessage writes a JSON-encoded message to the server's stdin.
func (c *LSPClient) sendMessage(msg interface{}) error {
	if c.shutdown {
		return fmt.Errorf("client is shutdown")
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// LSP messages use a header similar to HTTP: Content-Length followed by \r\n\r\n.
	content := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)
	_, err = c.stdin.Write([]byte(content))
	return err
}

// readMessages loops forever, parsing messages from the server's stdout.
func (c *LSPClient) readMessages() {
	reader := bufio.NewReader(c.stdout)

	for {
		if c.shutdown {
			return
		}

		// Parse the Content-Length header to know how many bytes to read next.
		contentLength := 0
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				break
			}

			var length int
			if n, _ := fmt.Sscanf(line, "Content-Length: %d", &length); n == 1 {
				contentLength = length
			}
		}

		if contentLength == 0 {
			continue
		}

		// Read the JSON body.
		buf := make([]byte, contentLength)
		_, err := io.ReadFull(reader, buf)
		if err != nil {
			return
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(buf, &msg); err != nil {
			continue
		}

		// If the message has an "id", it's a response to a request we sent.
		if idVal, hasID := msg["id"]; hasID {
			if c.logCallback != nil {
				c.logCallback("LSP", fmt.Sprintf("Received response with ID: %v (type: %T)", idVal, idVal))
			}
			if id, ok := idVal.(float64); ok {
				idInt := int64(id)
				if c.logCallback != nil {
					c.logCallback("LSP", fmt.Sprintf("Looking for response channel with ID=%d", idInt))
				}
				c.responseMutex.Lock()
				ch, exists := c.responses[idInt]
				if exists {
					if c.logCallback != nil {
						c.logCallback("LSP", fmt.Sprintf("Found channel for ID=%d, sending response", idInt))
					}
					delete(c.responses, idInt)
					c.responseMutex.Unlock()
					ch <- msg // Send response to the goroutine waiting for it.
				} else {
					if c.logCallback != nil {
						c.logCallback("LSP", fmt.Sprintf("No channel found for ID=%d", idInt))
					}
					c.responseMutex.Unlock()
				}
			} else {
				if c.logCallback != nil {
					c.logCallback("LSP", fmt.Sprintf("Failed to convert ID to int64: %v", idVal))
				}
			}
		}

		// If it has no "id", it's an asynchronous notification (like updated diagnostics).
		if _, hasID := msg["id"]; !hasID {
			c.handleNotification(msg)
		}
	}
}

// handleNotification processes messages initiated by the server.
func (c *LSPClient) handleNotification(msg map[string]interface{}) {
	method, ok := msg["method"].(string)
	if !ok {
		return
	}

	// Server is sending updated errors/warnings for the file.
	if method == "textDocument/publishDiagnostics" {
		params, ok := msg["params"].(map[string]interface{})
		if !ok {
			return
		}

		uri, _ := params["uri"].(string)
		if uri != c.uri {
			return
		}

		diagsRaw, ok := params["diagnostics"].([]interface{})
		if !ok {
			return
		}

		var diags []Diagnostic
		for _, d := range diagsRaw {
			diagJSON, _ := json.Marshal(d)
			var diag Diagnostic
			if json.Unmarshal(diagJSON, &diag) == nil {
				diags = append(diags, diag)
			}
		}

		c.diagMutex.Lock()
		c.diagnostics = diags
		c.diagMutex.Unlock()

		// Tell termbox to refresh the UI so signs appear in the gutter.
		termbox.Interrupt()
	}
}

// initialize sends the initial 'initialize' request to the server.
func (c *LSPClient) initialize() error {
	rootURI := "file://" + filepath.Dir(c.filename)
	params := map[string]interface{}{
		"processId": os.Getpid(),
		"rootUri":   rootURI,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"publishDiagnostics": map[string]interface{}{},
				"hover": map[string]interface{}{
					"contentFormat": []string{"plaintext"},
				},
				"completion": map[string]interface{}{
					"completionItem": map[string]interface{}{
						"snippetSupport": false,
					},
				},
			},
		},
	}

	if err := c.sendRequest("initialize", params); err != nil {
		return err
	}

	return c.sendNotification("initialized", map[string]interface{}{})
}

// sendDidOpen notifies the server that a file has been opened.
func (c *LSPClient) sendDidOpen(content string) error {
	languageID := strings.ToLower(c.fileType.Name)
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        c.uri,
			"languageId": languageID,
			"version":    1,
			"text":       content,
		},
	}
	return c.sendNotification("textDocument/didOpen", params)
}

// SendDidChange notifies the server of changes to the document content.
func (c *LSPClient) SendDidChange(content string) error {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":     c.uri,
			"version": c.nextID(),
		},
		"contentChanges": []interface{}{
			map[string]interface{}{
				"text": content,
			},
		},
	}
	return c.sendNotification("textDocument/didChange", params)
}

// GetDiagnostics returns a copy of the current file diagnostics.
func (c *LSPClient) GetDiagnostics() []Diagnostic {
	c.diagMutex.RLock()
	defer c.diagMutex.RUnlock()

	result := make([]Diagnostic, len(c.diagnostics))
	copy(result, c.diagnostics)
	return result
}

// Definition requests the location of the definition of the symbol at cursor.
func (c *LSPClient) Definition(line, character int) ([]Location, error) {
	id := c.nextID()
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": c.uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}

	responseChan := make(chan map[string]interface{}, 1)
	c.responseMutex.Lock()
	c.responses[id] = responseChan
	c.responseMutex.Unlock()

	if err := c.sendRequestWithID(id, "textDocument/definition", params); err != nil {
		c.responseMutex.Lock()
		delete(c.responses, id)
		c.responseMutex.Unlock()
		return nil, err
	}

	select {
	case resp := <-responseChan:
		if err, ok := resp["error"]; ok {
			return nil, fmt.Errorf("LSP error: %v", err)
		}

		result := resp["result"]
		if result == nil {
			return nil, nil
		}

		resJSON, _ := json.Marshal(result)

		// Definition can return a single Location or an array of them.
		var loc Location
		if err := json.Unmarshal(resJSON, &loc); err == nil && loc.URI != "" {
			return []Location{loc}, nil
		}

		var locs []Location
		if err := json.Unmarshal(resJSON, &locs); err == nil {
			return locs, nil
		}

		return nil, nil
	case <-time.After(5 * time.Second):
		c.responseMutex.Lock()
		delete(c.responses, id)
		c.responseMutex.Unlock()
		return nil, fmt.Errorf("LSP request timeout")
	}
}

// Hover requests documentation information for the symbol at cursor.
func (c *LSPClient) Hover(line, character int) (string, error) {
	id := c.nextID()
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": c.uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}

	responseChan := make(chan map[string]interface{}, 1)
	c.responseMutex.Lock()
	c.responses[id] = responseChan
	c.responseMutex.Unlock()

	if err := c.sendRequestWithID(id, "textDocument/hover", params); err != nil {
		c.responseMutex.Lock()
		delete(c.responses, id)
		c.responseMutex.Unlock()
		return "", err
	}

	select {
	case resp := <-responseChan:
		if err, ok := resp["error"]; ok {
			return "", fmt.Errorf("LSP error: %v", err)
		}

		result := resp["result"]
		if result == nil {
			return "", nil
		}

		// Hover responses are complex: they can be strings, objects, or arrays.
		resMap, ok := result.(map[string]interface{})
		if !ok {
			return "", nil
		}

		contents := resMap["contents"]
		if contents == nil {
			return "", nil
		}

		if mc, ok := contents.(map[string]interface{}); ok {
			if val, ok := mc["value"].(string); ok {
				return stripMarkdown(val), nil
			}
		}

		if s, ok := contents.(string); ok {
			return stripMarkdown(s), nil
		}

		if ss, ok := contents.([]interface{}); ok {
			var result strings.Builder
			for i, s := range ss {
				if str, ok := s.(string); ok {
					result.WriteString(stripMarkdown(str))
					if i < len(ss)-1 {
						result.WriteString("\n")
					}
				} else if m, ok := s.(map[string]interface{}); ok {
					if val, ok := m["value"].(string); ok {
						result.WriteString(stripMarkdown(val))
						if i < len(ss)-1 {
							result.WriteString("\n")
						}
					}
				}
			}
			return strings.TrimSpace(result.String()), nil
		}

		return "", nil
	case <-time.After(5 * time.Second):
		c.responseMutex.Lock()
		delete(c.responses, id)
		c.responseMutex.Unlock()
		return "", fmt.Errorf("LSP request timeout")
	}
}

// Completion requests a list of completion items for the symbol at cursor.
func (c *LSPClient) Completion(line, character int) ([]CompletionItem, error) {
	id := c.nextID()
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": c.uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}

	if c.logCallback != nil {
		c.logCallback("LSP", fmt.Sprintf("Requesting completion at %d:%d (ID=%d)", line, character, id))
	}

	responseChan := make(chan map[string]interface{}, 1)
	c.responseMutex.Lock()
	c.responses[id] = responseChan
	c.responseMutex.Unlock()

	if err := c.sendRequestWithID(id, "textDocument/completion", params); err != nil {
		c.responseMutex.Lock()
		delete(c.responses, id)
		c.responseMutex.Unlock()
		return nil, err
	}

	select {
	case resp := <-responseChan:
		if c.logCallback != nil {
			c.logCallback("LSP", fmt.Sprintf("Received completion response (ID=%d)", id))
		}
		if err, ok := resp["error"]; ok {
			return nil, fmt.Errorf("LSP error: %v", err)
		}

		result := resp["result"]
		if result == nil {
			return nil, nil
		}

		resJSON, _ := json.Marshal(result)

		// Completion can return a CompletionList or an array of CompletionItems.
		var compList CompletionList
		if err := json.Unmarshal(resJSON, &compList); err == nil {
			return compList.Items, nil
		}

		var compItems []CompletionItem
		if err := json.Unmarshal(resJSON, &compItems); err == nil {
			return compItems, nil
		}

		return nil, nil
	case <-time.After(10 * time.Second):
		if c.logCallback != nil {
			c.logCallback("LSP", fmt.Sprintf("Completion request timed out (ID=%d)", id))
		}
		c.responseMutex.Lock()
		delete(c.responses, id)
		c.responseMutex.Unlock()
		return nil, fmt.Errorf("LSP request timeout")
	}
}

// sendRequestWithID helper to send a request with a pre-generated ID.
func (c *LSPClient) sendRequestWithID(id int64, method string, params interface{}) error {
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}
	return c.sendMessage(request)
}

// Shutdown gracefully closes the LSP client and stops the server process.
func (c *LSPClient) Shutdown() {
	c.shutdownOnce.Do(func() {
		c.shutdown = true

		c.sendRequest("shutdown", nil)
		c.sendNotification("exit", nil)

		if c.stdin != nil {
			c.stdin.Close()
		}
		if c.stdout != nil {
			c.stdout.Close()
		}

		if c.cmd != nil && c.cmd.Process != nil {
			c.cmd.Wait()
		}
	})
}

// stripMarkdown provides a very naive way to remove markdown formatting from LSP responses.
func stripMarkdown(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	inCodeBlock := false
	for _, line := range lines {
		// Ignore code block markers.
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			result = append(result, line)
			continue
		}

		l := line
		l = strings.ReplaceAll(l, "**", "")
		l = strings.ReplaceAll(l, "__", "")

		// Naive link stripping: [text](url) -> text
		for {
			start := strings.Index(l, "[")
			end := strings.Index(l, "](")
			if start != -1 && end != -1 && end > start {
				closeParen := strings.Index(l[end:], ")")
				if closeParen != -1 {
					l = l[:start] + l[start+1:end] + l[end+closeParen+1:]
					continue
				}
			}
			break
		}

		l = strings.ReplaceAll(l, "`", "")
		result = append(result, l)
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}
