# Shellexa

Shellexa is a command-line tool designed to interpret natural language descriptions and fetch corresponding shell commands using local LLMs through the Ollama framework. It enhances productivity by allowing users to automate command retrieval using advanced language models without relying on cloud services.

## Features

- **Local AI Integration**: Leverages local installations of large language models via the oLLama framework to generate shell command suggestions from descriptions.
- **Easy Configuration**: Simple setup to connect with different models supported by oLLama.
- **Cross-Platform Compatibility**: Works on Linux, macOS, and Windows.

## Prerequisites

To use Shellexa, make sure you meet the following requirements:

- **Go Installed**: You need Go installed on your machine. [Download Go](https://golang.org/dl/).
- **oLLama Framework**: Shellexa uses the oLLama framework to run LLMs locally. Follow the installation instructions for your platform:
  - **macOS**: [Download oLLama for macOS](https://ollama.com/download/macos)
  - **Windows**: [Download oLLama for Windows](https://ollama.com/download/windows)
  - **Linux**:
    ```bash
    curl -fsSL https://ollama.com/install.sh | sh
    ```
  - **Docker**:
    ```bash
    docker pull ollama/ollama
    ```

## Installation

1. **Download Shellexa**:
   Visit the [Releases page](https://github.com/44za12/shellexa/releases) to download the latest version for your OS.

2. **Make Executable** (Linux and macOS):
   ```bash
   chmod +x ~/Downloads/shellexa
   ```

3. **Move to PATH** (Optional):
   ```bash
   mv ~/Downloads/shellexa /usr/local/bin/
   ```

## Configuration

Run the following command to configure Shellexa with your oLLama setup:

```bash
shellexa configure
```

You'll need to enter details such as the oLLama API URL and the model name you intend to use. These settings are stored in `.shellexa` in your home directory.

## Usage

To fetch a shell command using Shellexa:

```bash
shellexa "list all the files that ends with .go in the current directory"
```

After fetching the command:
- Type `e` to execute it.
- Type `a` to abort the operation.
- Type `r` to rethink the command.

## Building from Source

Clone the repository and build Shellexa:

```bash
git clone https://github.com/44za12/shellexa.git
cd shellexa
go build
```

## Contributing

We welcome contributions!

## License

Shellexa is distributed under the MIT License. See [LICENSE](LICENSE) for more details.