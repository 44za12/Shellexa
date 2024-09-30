package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	configFileName = "config.json"
)

type Config struct {
	APIKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) > 1 && os.Args[1] == "configure" {
		if err := configure(); err != nil {
			log.Fatalf("Configuration error: %v", err)
		}
		return
	}

	prompt := strings.Join(os.Args[1:], " ")
	if prompt == "" {
		fmt.Println("Usage: shellexa <natural language prompt>")
		fmt.Println("       shellexa configure")
		os.Exit(1)
	}

	if err := runConversation(prompt); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

func configure() error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your Gemini API Key: ")
	apiKey, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading API key: %w", err)
	}
	apiKey = strings.TrimSpace(apiKey)

	fmt.Print("Enter the model name (default is gemini-1.5-flash): ")
	modelName, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading model name: %w", err)
	}
	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	config := Config{APIKey: apiKey, ModelName: modelName}
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("error getting config directory: %w", err)
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	configPath := filepath.Join(configDir, configFileName)
	if err := os.WriteFile(configPath, configJSON, 0600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Println("Configuration saved successfully.")
	return nil
}

func loadConfig() (Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return Config{}, fmt.Errorf("error getting config directory: %w", err)
	}

	configPath := filepath.Join(configDir, configFileName)
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		return Config{}, fmt.Errorf("error parsing config file: %w", err)
	}

	return config, nil
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".shellexa"), nil
}

func runConversation(initialPrompt string) error {
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return fmt.Errorf("error creating Gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(config.ModelName)
	chat := model.StartChat()
	chat.History = []*genai.Content{
		{Role: "user", Parts: []genai.Part{genai.Text(getSystemInfo())}},
		{Role: "model", Parts: []genai.Part{genai.Text("Understood.")}},
	}

	if err := handleUserPrompt(ctx, chat, initialPrompt); err != nil {
		return err
	}

	return nil
}

func handleUserPrompt(ctx context.Context, chat *genai.ChatSession, prompt string) error {
	for {
		fullPrompt := fmt.Sprintf("Please provide only the shell command to %s. Do not include any explanations, markdown formatting, or backticks. The command should be directly executable in a shell.", prompt)
		resp, err := chat.SendMessage(ctx, genai.Text(fullPrompt))
		if err != nil {
			return fmt.Errorf("error sending message: %w", err)
		}

		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			return fmt.Errorf("no response generated")
		}

		suggestedCommand := strings.TrimSpace(string(resp.Candidates[0].Content.Parts[0].(genai.Text)))

		fmt.Printf("Suggested command: %s\n", suggestedCommand)
		fmt.Print("Options: [e]xecute, [a]bort, [r]ethink: ")
		reader := bufio.NewReader(os.Stdin)
		choice, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading user input: %w", err)
		}
		choice = strings.TrimSpace(choice)

		switch choice {
		case "e":
			cmd := exec.Command("sh", "-c", suggestedCommand)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				errorMsg := fmt.Sprintf("Command failed: %s", err)
				fmt.Println(errorMsg)
				prompt = fmt.Sprintf("The command '%s' failed with error: %s. Please suggest a corrected command.", suggestedCommand, err)
			} else {
				fmt.Println("Command executed successfully.")
				return nil
			}
		case "a":
			fmt.Println("Command aborted.")
			return nil
		case "r":
			prompt = "Please provide an alternative command."
		default:
			fmt.Println("Invalid choice. Please try again.")
			prompt = "The user made an invalid choice. Please provide the same command again."
		}
	}
}

func getSystemInfo() string {
	var info strings.Builder
	info.WriteString("System Information:\n")
	info.WriteString(fmt.Sprintf("OS: %s\n", runtime.GOOS))
	info.WriteString(fmt.Sprintf("Architecture: %s\n", runtime.GOARCH))

	hostname, err := os.Hostname()
	if err == nil {
		info.WriteString(fmt.Sprintf("Hostname: %s\n", hostname))
	}

	currentDir, err := os.Getwd()
	if err == nil {
		info.WriteString(fmt.Sprintf("Current Directory: %s\n", currentDir))
	}
	info.WriteString("\nPlease keep this system information in mind when suggesting commands. Respond with 'Understood' if you acknowledge this information.")
	return info.String()
}
