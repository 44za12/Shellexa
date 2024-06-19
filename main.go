package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
)

type Config struct {
	APIURL string `json:"api_url"`
	Model  string `json:"model"`
}

const configFileName = "config.json"

type APIMessage struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIResponse struct {
	Message Message `json:"message"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: shellexa [configure | \"command\"]")
		os.Exit(1)
	}
	configPath := getConfigFilePath()
	scanner := bufio.NewScanner(os.Stdin)
	userInput := os.Args[1]

	if userInput == "configure" {
		configure(configPath)
		return
	}
	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Println("Error getting config, ensure you have ran `shellexa configure` before first use.\nError:", err)
		return
	}
	commandToExecute, err := fetchCommandFromAPI(config, userInput, "")
	if err != nil {
		fmt.Println("Error fetching command from API:", err)
		return
	}

	for {
		fmt.Println("Suggested command:", commandToExecute)
		fmt.Println("Options: [e] execute, [a] abort, [r] rethink")
		fmt.Println()

		if scanner.Scan() {
			userChoice := scanner.Text()
			switch userChoice {
			case "e":
				fmt.Print("\033[1A\033[K")
				if err := executeCommand(commandToExecute); err != nil {
					fmt.Println("Failed to execute command:", err)
					continue
				} else {
					return
				}
			case "a":
				fmt.Println("Operation aborted.")
				return
			case "r":
				fmt.Print("\033[1A\033[K")
				fmt.Println("Re-thinking the command...")
				if newCommand, err := rethinkCommand(config, userInput, commandToExecute, ""); err == nil {
					commandToExecute = newCommand // Update the command to the newly thought command
				} else {
					fmt.Println("Error during re-thinking:", err)
					return
				}
			default:
				fmt.Println("Invalid option. Please choose [e] execute, [a] abort, or [r] rethink.")
			}
		}
	}
}

func getConfigFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".shellexa", configFileName)
}

func configure(configPath string) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter API URL: ")
	scanner.Scan()
	apiURL := scanner.Text()

	fmt.Print("Enter Model Name: ")
	scanner.Scan()
	modelName := scanner.Text()

	config := Config{
		APIURL: apiURL,
		Model:  modelName,
	}
	if err := saveConfig(configPath, &config); err != nil {
		fmt.Printf("Error saving configuration: %s\n", err)
		return
	}
	fmt.Println("Configuration saved successfully.")
}

func saveConfig(path string, config *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return nil
}

func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

func fetchCommandFromAPI(config *Config, input string, errorMessage string) (string, error) {
	url := config.APIURL
	contextInfo := fmt.Sprintf("System: %s, Arch: %s", runtime.GOOS, runtime.GOARCH)
	prompt := fmt.Sprintf("Generate a shell command to achieve this: '%s'. System context: %s. Previous error (if any): '%s'. Provide only the command nothing else, not even any explanations or notes or suggestions, it is essential as this command would be directly executed.", input, contextInfo, errorMessage)
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		reqBody := APIMessage{
			Model: config.Model,
			Messages: []Message{
				{Role: "user", Content: prompt},
			},
			Stream: false,
		}
		jsonData, _ := json.Marshal(reqBody)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return "", err
		}

		commandToExecute := parseCommand(apiResp.Message.Content)
		if commandToExecute != "" {
			return commandToExecute, nil
		}
	}

	return "", fmt.Errorf("failed to retrieve a valid command after %d attempts", maxRetries)
}

func parseCommand(response string) string {
	re := regexp.MustCompile("(?s)```(?:[a-zA-Z]+\\s+)?(.*?)```|`([^`]+)`")
	matches := re.FindStringSubmatch(response)

	if len(matches) > 1 {
		for _, match := range matches[1:] {
			if match != "" {
				return match
			}
		}
	}
	return ""
}

func rethinkCommand(config *Config, initialComment, previousCommand, errorFeedback string) (string, error) {
	return fetchCommandFromAPI(config, initialComment, fmt.Sprintf("Failed to execute: %s. Error: %s.", previousCommand, errorFeedback))
}

func executeCommand(cmd string) error {
	commandParts := exec.Command("sh", "-c", cmd)
	commandOutput, err := commandParts.CombinedOutput()
	if err != nil {
		fmt.Println("Execution error:", err)
		fmt.Println(string(commandOutput))
		return err
	}
	fmt.Println(string(commandOutput))
	return nil
}
