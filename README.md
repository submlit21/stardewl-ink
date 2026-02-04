# Stardewl-Ink 🌀

Stardew Valley multiplayer tool using WebRTC for P2P connections, no port forwarding required.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![WebRTC](https://img.shields.io/badge/WebRTC-P2P-blue)](https://webrtc.org/)

## ✨ Features

- 🚀 **WebRTC P2P Connection** - Pair using connection codes, no port forwarding or complex configuration
- 🔗 **Simple Pairing System** - Host generates connection code, client enters to connect
- 📁 **Smart Mod Checking** - Automatically scans and compares Mod files, shows differences
- 🛠️ **Truly Cross-Platform** - Core in Go, native UI for each platform
- 🔒 **Privacy First** - No accounts, no login, no friend system
- ⚡ **High Performance** - Based on Pion WebRTC, stable and efficient P2P connections

## 🚀 Quick Start

### 1. Build from Source

```bash
# Clone the repository
git clone git@github.com:submlit21/stardewl-ink.git
cd stardewl-ink

# Download dependencies
go mod download

# Build the project
make build

# Or build for all platforms
make cross-build-all
```

### 2. Run Signaling Server

You need a signaling server for WebRTC handshake. Run it on any machine:

```bash
# On your server or local machine
./dist/stardewl-signaling
# Server starts on port 8080
```

### 3. Connect Players

**Player 1 (Host):**
```bash
./dist/stardewl --host --timeout=300
# Output: Connection code: 123456
```

**Player 2 (Client):**
```bash
./dist/stardewl --join=123456 --timeout=300
```

## 📦 Pre-built Binaries

Check [Releases](https://github.com/submlit21/stardewl-ink/releases) for pre-built binaries for:
- Windows (.exe)
- macOS (Apple Silicon/Intel)
- Linux

## 🛠️ Command Line Usage

```bash
# Host mode (creates a room)
./dist/stardewl --host [--timeout=SECONDS] [--signaling=URL]

# Client mode (joins a room)
./dist/stardewl --join=CODE [--timeout=SECONDS] [--signaling=URL]

# List Mods
./dist/stardewl --list-mods [--mods=PATH]

# Interactive mode
./dist/stardewl --interactive

# Help
./dist/stardewl --help
```

**Options:**
- `--timeout`: Timeout in seconds (0 = wait indefinitely, default: 0)
- `--signaling`: Signaling server URL (default: ws://localhost:8080/ws)
- `--mods`: Mods folder path (default: auto-detect)
- `--verbose`: Enable verbose logging

## 🏗️ Project Structure

```
stardewl-ink/
├── core/                 # Core WebRTC connection library
│   ├── connection.go    # WebRTC connection management
│   ├── mods.go         # Mod file scanning and comparison
│   ├── messages.go     # Message protocol definitions
│   └── core.go         # Client main logic
├── signaling/           # Signaling server
│   └── main.go         # WebSocket signaling server
├── cmd/stardewl/        # Command line interface
├── examples/           # Example code
└── dist/              # Build outputs
```

## 🔧 Development

### Build Commands

```bash
# Build for current platform
make build

# Build for specific platforms
make cross-build-windows  # Windows .exe files
make cross-build-macos    # macOS binaries
make cross-build-linux    # Linux binaries

# Build for all platforms
make cross-build-all

# Clean build artifacts
make clean
```

### Dependencies

- Go 1.22+
- [Pion WebRTC](https://github.com/pion/webrtc) - WebRTC implementation
- Standard Go libraries

## 🔒 Privacy & Security

- **No Data Collection**: All connections are direct P2P
- **No Accounts**: No registration or login required
- **Local First**: Mod scanning happens locally
- **Encrypted**: WebRTC provides end-to-end encryption

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/submlit21/stardewl-ink/issues)
- **Questions**: Open an issue or discussion

---

**Happy farming!** 🌾