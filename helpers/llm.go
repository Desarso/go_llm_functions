package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strings"
)

// ChatRequest represents the request body for the API
var DEFAULT_URL = "https://openrouter.ai/api/v1/chat/completions"
var API_KEY = ""
var MODEL = "openai/gpt-4o-mini"

// extractParamNames extracts parameter names from runtime function info
func extractParamNames(fn interface{}) []string {
	// Get function type
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil
	}

	// Get function value
	fnValue := reflect.ValueOf(fn)

	// Get function pointer
	ptr := fnValue.Pointer()

	// Get runtime function info
	rf := runtime.FuncForPC(ptr)
	if rf == nil {
		return nil
	}

	// Try to find source location
	file, line := rf.FileLine(ptr)
	if file == "" {
		return nil
	}

	// Read the source file
	src, err := os.ReadFile(file)
	if err != nil {
		return nil
	}

	// Find the line containing the function
	lines := strings.Split(string(src), "\n")
	if line-1 >= len(lines) {
		return nil
	}

	// Extract parameter names from function declaration
	funcLine := lines[line-1]
	start := strings.Index(funcLine, "(")
	end := strings.Index(funcLine, ")")
	if start == -1 || end == -1 {
		return nil
	}

	// Parse parameters
	params := strings.Split(funcLine[start+1:end], ",")
	var paramNames []string
	for _, p := range params {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Extract name from "name type" format
		parts := strings.Fields(p)
		if len(parts) > 0 {
			paramNames = append(paramNames, parts[0])
		}
	}

	return paramNames
}

// Options represents configuration options for the LLM function
type Options struct {
	Debug bool
	Top   int
}

// ToolFunction stores a function that can be called by the LLM
type ToolFunction struct {
	fn         func(map[string]interface{}) (string, error)
	parameters *Parameters
}

// Map to store tool functions
var toolFunctions = make(map[string]*ToolFunction)

// CreateTool creates a function tool that can be used by the LLM
func CreateTool(name string, description string, fn interface{}) *Tool {
	// Get the function type
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("fn must be a function")
	}

	// Create parameter schema based on function parameters
	parameters := &Parameters{
		Type:       "object",
		Properties: make(map[string]*Field),
		Required:   make([]string, 0),
	}

	// Extract parameter names from function
	if paramNames := extractParamNames(fn); len(paramNames) == fnType.NumIn() {
		// Add each function parameter to the schema with actual names
		for i := 0; i < fnType.NumIn(); i++ {
			paramType := fnType.In(i)
			paramName := paramNames[i]
			parameters.Properties[paramName] = &Field{
				Description: fmt.Sprintf("The %s parameter of type %v", paramName, paramType.Kind()),
				Type:        getJSONType(paramType.Kind()),
			}
			parameters.Required = append(parameters.Required, paramName)
		}
	} else {
		// Fallback to generic names if we can't get the actual names
		for i := 0; i < fnType.NumIn(); i++ {
			paramType := fnType.In(i)
			paramName := fmt.Sprintf("param%d", i)
			parameters.Properties[paramName] = &Field{
				Description: fmt.Sprintf("Parameter %d of type %v", i+1, paramType.Kind()),
				Type:        getJSONType(paramType.Kind()),
			}
			parameters.Required = append(parameters.Required, paramName)
		}
	}

	// Create wrapper function that handles argument conversion
	wrapper := func(args map[string]interface{}) (string, error) {
		// Create slice to hold converted arguments
		fnArgs := make([]reflect.Value, fnType.NumIn())

		// Get parameter names
		paramNames := extractParamNames(fn)

		// Convert each argument to the correct type
		for i := 0; i < fnType.NumIn(); i++ {
			paramType := fnType.In(i)

			// Try actual name first, fall back to generic name
			var paramName string
			if i < len(paramNames) {
				paramName = paramNames[i]
			} else {
				paramName = fmt.Sprintf("param%d", i)
			}

			// Get argument value
			arg, ok := args[paramName]
			if !ok {
				return "", fmt.Errorf("missing parameter: %s", paramName)
			}

			// Convert argument to the correct type
			converted := reflect.ValueOf(arg)
			if converted.Type().ConvertibleTo(paramType) {
				fnArgs[i] = converted.Convert(paramType)
			} else {
				return "", fmt.Errorf("cannot convert parameter %s to type %v", paramName, paramType)
			}
		}

		// Call the function with converted arguments
		result := reflect.ValueOf(fn).Call(fnArgs)
		if len(result) != 1 {
			return "", fmt.Errorf("function must return exactly one value")
		}

		return result[0].String(), nil
	}

	// Store the wrapper function
	toolFunctions[name] = &ToolFunction{
		fn:         wrapper,
		parameters: parameters,
	}

	return &Tool{
		Name:        name,
		Type:        "function",
		Description: description,
		Function: &Function{
			Name:       name,
			Parameters: parameters,
		},
	}
}

// getJSONType converts Go types to JSON schema types
func getJSONType(k reflect.Kind) string {
	switch k {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	default:
		return "string"
	}
}

// ExecuteTool executes a stored tool function
func ExecuteTool(name string, arguments string) (string, error) {
	tool, exists := toolFunctions[name]
	if !exists {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	// Parse the arguments JSON into a map
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("error parsing arguments: %w", err)
	}

	// Execute the function with the parsed arguments
	return tool.fn(args)
}

func LLM(fn func(string) string, opts ...interface{}) func(string) string {
	// Parse optional parameters
	var systemMessage string
	var options Options
	var tools []*Tool
	for _, opt := range opts {
		switch v := opt.(type) {
		case string:
			systemMessage = v
		case Options:
			options = v
		case *Tool:
			tools = append(tools, v)
		case []*Tool:
			tools = append(tools, v...)
		}
	}

	return func(input string) string {
		// Get the original function result
		original := fn(input)

		// Create messages array
		messages := []Message{}

		// Add system message if provided
		if systemMessage != "" {
			messages = append(messages, Message{
				Role:    "system",
				Content: systemMessage,
			})
		}

		// Add user message
		messages = append(messages, Message{
			Role:    "user",
			Content: original,
		})

		// Send the chat request with tools if provided
		response, err := chat(API_KEY, MODEL, messages, options, tools...)
		if err != nil {
			if options.Debug {
				fmt.Printf("Error in LLM request: %v\n", err)
			}
			return original
		}

		// Handle the response
		if len(response.Choices) == 0 {
			if options.Debug {
				fmt.Println("No response from LLM")
			}
			return original
		}

		// Check for tool calls in the response
		if len(response.Choices) > 0 && response.Choices[0].Message != nil && len(response.Choices[0].Message.ToolCalls) > 0 {
			// Execute each tool call
			for _, toolCall := range response.Choices[0].Message.ToolCalls {
				if toolCall.Function != nil {
					// Execute the tool and get result
					result, err := ExecuteTool(toolCall.Function.Name, toolCall.Function.Arguments)
					if err != nil {
						if options.Debug {
							fmt.Printf("Error executing tool %s: %v\n", toolCall.Function.Name, err)
						}
						continue
					}

					// Add the assistant message with tool calls
					messages = append(messages, Message{
						Role:      "assistant",
						Content:   "",
						ToolCalls: []*ToolCall{toolCall},
					})

					// Add the tool result message with tool_call_id
					messages = append(messages, Message{
						Role:       "tool",
						Content:    result,
						ToolCallID: toolCall.Id,
					})
				}
			}

			// Get final response with tool results
			finalResponse, err := chat(API_KEY, MODEL, messages, options)
			if err != nil {
				if options.Debug {
					fmt.Printf("Error in final LLM request: %v\n", err)
				}
				return original
			}

			if len(finalResponse.Choices) > 0 {
				return finalResponse.Choices[0].Message.Content
			}
		}

		// Return the original response if no tool calls
		return response.Choices[0].Message.Content
	}
}

func chat(apiKey, model string, messages []Message, options Options, tools ...*Tool) (*ResponseData, error) {
	// Prepare the request payload
	requestBody := Request{
		Model:    model,
		Messages: messages,
	}

	// Add tools if provided
	if len(tools) > 0 {
		toolSlice := make([]Tool, len(tools))
		for i, t := range tools {
			if t != nil {
				toolSlice[i] = *t
			}
		}
		requestBody.Tools = toolSlice
		requestBody.ToolChoice = "auto"
	}

	// Convert the struct to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	// Print the raw request in JSON
	if options.Debug {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, jsonData, "", "  "); err != nil {
			fmt.Printf("Error formatting JSON: %v\n", err)
		} else {
			fmt.Printf("Raw request: %s\n", prettyJSON.String())
		}
	}

	// Create a new HTTP request
	url := DEFAULT_URL
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Print the raw response in JSON
	body = bytes.TrimSpace(body)
	if options.Debug {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
			fmt.Printf("Error formatting JSON: %v\n", err)
		} else {
			fmt.Printf("Raw response: %s\n", prettyJSON.String())
		}
	}

	// Declare chatResponse variable
	var chatResponse ResponseData
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	return &chatResponse, nil
}

func chatStream(messages []Message) (<-chan string, <-chan error) {
	// Create channels for chunks and errors
	chunks := make(chan string)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		// Prepare the request payload
		requestBody := Request{
			Model:    MODEL,
			Messages: messages,
			Stream:   true,
		}

		// Convert the struct to JSON
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			errs <- fmt.Errorf("error marshaling request body: %w", err)
			return
		}

		// Create a new HTTP request
		url := DEFAULT_URL
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			errs <- fmt.Errorf("error creating request: %w", err)
			return
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+API_KEY)

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("error sending request: %w", err)
			return
		}
		defer resp.Body.Close()

		// Check for non-200 status codes
		if resp.StatusCode != http.StatusOK {
			errs <- fmt.Errorf("request failed with status code: %d", resp.StatusCode)
			return
		}

		// Stream the response line by line
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			chunks <- scanner.Text() // Send each line to the channel
		}

		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			errs <- fmt.Errorf("error reading streamed response: %w", err)
		}
	}()

	return chunks, errs
}
