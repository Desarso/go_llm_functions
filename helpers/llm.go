package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// ChatRequest represents the request body for the API

var GROQ_API_KEY = ""
var MODEL = "llama-3.3-70b-versatile"

// func init() {
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatalf("Error loading .env file: %v", err)
// 	}

// 	GROQ_API_KEY = os.Getenv("GROQ_API_KEY")

// }

func SingleResponse(message string) (string, error) {
	// Define the model and messages
	messages := []Message{
		{
			Role:    "user",
			Content: message,
		},
	}

	// Send the chat request
	response, err := chat(GROQ_API_KEY, MODEL, messages)
	if err != nil {
		return "", fmt.Errorf("error sending chat request: %w", err)
	}

	// Return the response message
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response message found")
	}
	// fmt.Println("res", response.Choices[0].Message)
	return response.Choices[0].Message.Content, nil
}

func chat(apiKey, model string, messages []Message) (*ResponseData, error) {
	// Prepare the request payload
	requestBody := Request{
		Model:    model,
		Messages: messages,
	}

	// Convert the struct to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	// Create a new HTTP request
	url := "https://api.groq.com/openai/v1/chat/completions"
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

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	// Read the body and convert to string
	// body, err := ioutil.ReadAll(resp.Body) // Or use io.ReadAll in Go 1.16+
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Print the body as a string
	// fmt.Println(string(body))

	// Parse the response
	var chatResponse ResponseData
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// fmt.Println(chatResponse.Choices)

	return &chatResponse, nil
}

func chatStream(messages []Message) (<-chan string, <-chan error) {
	// Create channels for chunks and errors
	chunks := make(chan string)
	errs := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errs)

		//make a greet user tool
		// tools := []Tool{
		// 	{
		// 		Name:        "greet_user",
		// 		Type:        "function",
		// 		Description: "Generates a greeting for a user based on their name.",
		// 		Function: &Function{
		// 			Name: "greet_user_function",
		// 			Parameters: &Parameters{
		// 				Type: "object",
		// 				Properties: map[string]*Field{
		// 					"name": {
		// 						Description: "The name of the user.",
		// 						Type:        "string",
		// 					},
		// 				},
		// 				Required: []string{"name"},
		// 			},
		// 		},
		// 	},
		// }
		// Prepare the request payload
		requestBody := Request{
			Model:    MODEL,
			Messages: messages,
			Stream:   true,
			// ToolChoice: "auto",
			// Tools:      tools,
		}

		// Convert the struct to JSON
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			errs <- fmt.Errorf("error marshaling request body: %w", err)
			return
		}

		// Create a new HTTP request
		url := "https://api.groq.com/openai/v1/chat/completions"
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			errs <- fmt.Errorf("error creating request: %w", err)
			return
		}

		//print request body
		// fmt.Println(string(jsonData))

		// Add headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+GROQ_API_KEY)

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			errs <- fmt.Errorf("error sending request: %w", err)
			return
		}
		defer resp.Body.Close()

		// fmt.Println("Response Status:", resp)

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

		//when the scanner is done add the last message, print last chunk
	}()

	return chunks, errs
}

func Ell(task func(string) string) func(string) string {
	return func(arg string) string {
		original := task(arg)
		result, err := SingleResponse(original)
		if err != nil {
			fmt.Println(err)
		}
		return result

	}
}
