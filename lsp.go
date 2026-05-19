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
	writeMutex    sync.Mutex // Protects concurrent writes to stdin.
	fileType      *FileType  // Associated file type for language ID.
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
	Label         string               `json:"label"`
	LabelDetails  *CompletionItemLabel `json:"labelDetails"`
	Kind          int                  `json:"kind"`
	Detail        string               `json:"detail"`
	Documentation interface{}          `json:"documentation"`
	InsertText    string               `json:"insertText"`
	FilterText    string               `json:"filterText"`
	TextEdit      *TextEdit            `json:"textEdit"`
	Data          interface{}          `json:"data"` // Opaque data for resolve request
}

type CompletionItemLabel struct {
	Detail      string `json:"detail"`      // e.g. (int a, int b)
	Description string `json:"description"` // e.g. int
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
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

	// Suppress the server's own internal log messages (stderr) and redirect to logCallback.
	stderr, err := client.cmd.StderrPipe()
	if err == nil {
		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				if client.logCallback != nil {
					client.logCallback("LSP-stderr", scanner.Text())
				}
			}
		}()
	}

	client.stdin, err = client.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	client.stdout, err = client.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Set working directory to the project root to help server find config files and resolve relative paths.
	client.cmd.Dir = findProjectRoot(absPath)

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

// Request sends a JSON-RPC request and waits for a response (up to 5s).
func (c *LSPClient) Request(method string, params interface{}) (map[string]interface{}, error) {
	id := c.nextID()
	responseChan := make(chan map[string]interface{}, 1)
	c.responseMutex.Lock()
	c.responses[id] = responseChan
	c.responseMutex.Unlock()

	if err := c.sendRequestWithID(id, method, params); err != nil {
		c.responseMutex.Lock()
		delete(c.responses, id)
		c.responseMutex.Unlock()
		return nil, err
	}

	select {
	case resp := <-responseChan:
		if errVal, ok := resp["error"]; ok {
			return nil, fmt.Errorf("LSP error: %v", errVal)
		}
		return resp, nil
	case <-time.After(10 * time.Second):
		c.responseMutex.Lock()
		delete(c.responses, id)
		c.responseMutex.Unlock()
		return nil, fmt.Errorf("LSP request timeout: %s", method)
	}
}

// sendRequest sends a JSON-RPC request and expects a response.
func (c *LSPClient) sendRequest(method string, params interface{}) error {
	_, err := c.Request(method, params)
	return err
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

	if c.logCallback != nil {
		msgStr := string(data)
		if len(msgStr) > 500 {
			msgStr = msgStr[:500] + "..."
		}
		c.logCallback("LSP-send", msgStr)
	}

	// LSP messages use a header similar to HTTP: Content-Length followed by \r\n\r\n.
	content := fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(data), data)

	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
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

			lowerLine := strings.ToLower(line)
			var length int
			if strings.HasPrefix(lowerLine, "content-length:") {
				if n, _ := fmt.Sscanf(lowerLine, "content-length: %d", &length); n == 1 {
					contentLength = length
				}
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

		// If the message has an "id", it's either a response to a request we sent
		// or a request from the server to us.
		if idVal, hasID := msg["id"]; hasID {
			method, isServerRequest := msg["method"].(string)

			if isServerRequest {
				if c.logCallback != nil {
					c.logCallback("LSP", fmt.Sprintf("Received server request: %s (ID: %v)", method, idVal))
				}
				// Handle server-to-client requests.
				c.handleServerRequest(method, idVal, msg["params"])
			} else {
				if c.logCallback != nil {
					c.logCallback("LSP", fmt.Sprintf("Received response for ID: %v", idVal))
				}

				var idInt int64
				validID := false
				switch v := idVal.(type) {
				case float64:
					idInt = int64(v)
					validID = true
				case string:
					fmt.Sscanf(v, "%d", &idInt)
					validID = true
				}

				if validID {
					c.responseMutex.Lock()
					ch, exists := c.responses[idInt]
					if exists {
						delete(c.responses, idInt)
						c.responseMutex.Unlock()
						ch <- msg
					} else {
						if c.logCallback != nil {
							c.logCallback("LSP", fmt.Sprintf("No channel found for ID=%d", idInt))
						}
						c.responseMutex.Unlock()
					}
				}
			}
		}

		// If it has no "id", it's an asynchronous notification.
		if _, hasID := msg["id"]; !hasID {
			c.handleNotification(msg)
		}
	}
}

// handleServerRequest responds to requests initiated by the server.
func (c *LSPClient) handleServerRequest(method string, id interface{}, params interface{}) {
	// For now, we provide minimal responses to keep the server happy.
	var result interface{} = nil

	if method == "workspace/configuration" {
		// Return empty settings for any requested scope.
		if p, ok := params.(map[string]interface{}); ok {
			if items, ok := p["items"].([]interface{}); ok {
				res := make([]interface{}, len(items))
				for i := range res {
					res[i] = map[string]interface{}{}
				}
				result = res
			}
		}
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	c.sendMessage(response)
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

// findProjectRoot looks for a project root marker like .git, compile_commands.json, or .clangd.
func findProjectRoot(path string) string {
	dir := filepath.Dir(path)
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "compile_commands.json")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, ".clangd")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Dir(path)
}

// initialize sends the initial 'initialize' request to the server.
func (c *LSPClient) initialize() error {
	rootPath := findProjectRoot(c.filename)
	rootURI := "file://" + rootPath
	params := map[string]interface{}{
		"processId": os.Getpid(),
		"rootUri":   rootURI,
		"rootPath":  rootPath, // Deprecated but some servers still use it.
		"workspaceFolders": []map[string]interface{}{
			{
				"uri":  rootURI,
				"name": filepath.Base(rootPath),
			},
		},
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"synchronization": map[string]interface{}{
					"didSave":             true,
					"dynamicRegistration": false,
					"willSave":            false,
					"willSaveWaitUntil":   false,
				},
				"publishDiagnostics": map[string]interface{}{},
				"hover": map[string]interface{}{
					"contentFormat": []string{"plaintext"},
				},
				"completion": map[string]interface{}{
					"completionItem": map[string]interface{}{
						"snippetSupport":          false,
						"resolveSupport":          map[string]interface{}{"properties": []string{"documentation", "detail"}},
						"insertReplaceSupport":    true,
						"labelDetailsSupport":     true,
						"deprecatedSupport":       true,
						"commitCharactersSupport": false,
					},
					"contextSupport": true,
				},
				"definition": map[string]interface{}{
					"dynamicRegistration": false,
					"linkSupport":         false,
				},
			},
			"workspace": map[string]interface{}{
				"configuration":    true,
				"workspaceFolders": true,
			},
		},
	}

	// Move textDocumentSync to top level of capabilities if needed by some servers,
	// though it's technically under textDocument in some versions.
	// Actually, the spec says it should be under capabilities for server capabilities,
	// but for client capabilities it is under textDocument.
	// However, many servers like gopls prefer it at a certain location.
	// Let's add it to the top level of capabilities as well just in case.
	params["capabilities"].(map[string]interface{})["textDocumentSync"] = 1 // Full

	if err := c.sendRequest("initialize", params); err != nil {
		return err
	}

	return c.sendNotification("initialized", map[string]interface{}{})
}

// sendDidOpen notifies the server that a file has been opened.
func (c *LSPClient) sendDidOpen(content string) error {
	languageID := c.fileType.LanguageID
	if languageID == "" {
		languageID = strings.ToLower(c.fileType.Name)
	}
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
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": c.uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}

	resp, err := c.Request("textDocument/definition", params)
	if err != nil {
		return nil, err
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
}

// Hover requests documentation information for the symbol at cursor.
func (c *LSPClient) Hover(line, character int) (string, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": c.uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
	}

	resp, err := c.Request("textDocument/hover", params)
	if err != nil {
		return "", err
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
}

// ResolveCompletion requests additional details for a completion item.
func (c *LSPClient) ResolveCompletion(item CompletionItem) (CompletionItem, error) {
	resp, err := c.Request("completionItem/resolve", item)
	if err != nil {
		return item, err
	}

	result := resp["result"]
	if result == nil {
		return item, nil
	}

	resJSON, _ := json.Marshal(result)
	var resolvedItem CompletionItem
	if err := json.Unmarshal(resJSON, &resolvedItem); err != nil {
		return item, err
	}

	return resolvedItem, nil
}

// getDocumentationString extracts a plain string from the Documentation field.
func (c *LSPClient) getDocumentationString(doc interface{}) string {
	if doc == nil {
		return ""
	}
	if s, ok := doc.(string); ok {
		return stripMarkdown(s)
	}
	if m, ok := doc.(map[string]interface{}); ok {
		if val, ok := m["value"].(string); ok {
			return stripMarkdown(val)
		}
	}
	return ""
}

// Completion requests a list of completion items for the symbol at cursor.
func (c *LSPClient) Completion(line, character int) ([]CompletionItem, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": c.uri,
		},
		"position": map[string]interface{}{
			"line":      line,
			"character": character,
		},
		"context": map[string]interface{}{
			"triggerKind": 1, // Invited
		},
	}

	resp, err := c.Request("textDocument/completion", params)
	if err != nil {
		return nil, err
	}

	result := resp["result"]
	if result == nil {
		return nil, nil
	}

	resJSON, _ := json.Marshal(result)

	// Completion can return a CompletionList or an array of CompletionItems.
	// Try unmarshaling into array of items first as it's more direct.
	var compItems []CompletionItem
	if err := json.Unmarshal(resJSON, &compItems); err == nil && compItems != nil {
		return compItems, nil
	}

	var compList CompletionList
	if err := json.Unmarshal(resJSON, &compList); err == nil {
		return compList.Items, nil
	}

	return nil, nil
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
