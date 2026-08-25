# ⚡ VibeTerm — Modern Infrastructure Workspace & Terminal

> **High-Performance, Lightweight Infrastructure Workspace, SSH Manager, and Terminal Multiplexer powered by Go, Wails, React, and Xterm.js.**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)](https://react.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org)
[![TailwindCSS](https://img.shields.io/badge/TailwindCSS-v3-38B2AC?style=flat&logo=tailwind-css)](https://tailwindcss.com)
[![Wails](https://img.shields.io/badge/Wails-v2-DF1A2A?style=flat)](https://wails.io)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 🌟 Why VibeTerm?

Traditional server management tools force a painful trade-off:
- **Electron-based terminals** consume 1GB+ of RAM, feel sluggish, and introduce typing latency.
- **Legacy native GUIs** suffer from rigid styling, limited keyboard ergonomics, and clunky interfaces.

**VibeTerm** bridges the gap. By pairing a high-performance **Go backend** with **Wails v2** and a sleek **React/Tailwind/Xterm.js frontend**, VibeTerm delivers a gorgeous, VS Code / Cursor-grade monochrome user experience while maintaining a tiny memory footprint (< 50MB) and instant startup times.

---

## ✨ Key Features

- **🖥️ Full Server & Endpoint Management**: Complete CRUD dashboard for managing SSH servers across environments (Production, Staging, Dev) with tagging, folders, and operator notes.
- **⚡ Interactive Terminal Engine**: Powered by `@xterm/xterm` with true native PTY streaming for both local shells (bash/zsh) and remote SSH sessions.
- **📋 Ergonomic Terminal Context Menu**: Full right-click support for Copy, Paste, Select All, and Clear Terminal.
- **🛡️ Multi-Hop Bastion (Jump Host) Support**: Connect across complex bastion server chains with independent credentials per hop.
- **🔀 Port Forwarding & SOCKS5 Tunneling**: Manage Local (`-L`), Remote (`-R`), and Dynamic SOCKS5 (`-D`) proxies with live traffic metrics.
- **🔒 Enterprise Security Vault**: Credentials and private keys stored safely with local **AES-256-GCM** encryption.
- **🎨 High-Contrast Monochrome Aesthetic**: Minimalist, distraction-free dark theme with adaptive system theme switching and high-contrast syntax highlighting.

---

## 🏛️ Architecture

```
vibeterm/
├── app.go                         # Wails Go application bridge & controller
├── main.go                        # Application entry point & native window config
├── internal/
│   ├── models/                    # Shared data structures (Host, Tunnels, Keys)
│   ├── config/                    # AES-256-GCM Vault & thread-safe config persistence
│   ├── ssh/                       # Bastion jump-host manager & interactive SSH sessions
│   ├── pty/                       # Native OS PTY process multiplexer (bash/zsh)
│   ├── forwarding/                # Port forwarding (-L, -R, -D SOCKS5 proxy engine)
│   ├── scanner/                   # Subnet CIDR scanner & live health probe mesh
│   ├── gitops/                    # Git repository configuration sync engine
│   └── automation/                # Trigger macros & automation scripts
├── frontend/                      # React 18 + TypeScript + TailwindCSS Web Client
│   ├── src/
│   │   ├── App.tsx                # Unified header, activity bar, and workspace layout
│   │   ├── TerminalComponent.tsx  # Live Xterm.js canvas with context menu integration
│   │   ├── HostManagerView.tsx    # Server grid dashboard with filters & quick connect
│   │   ├── HostModal.tsx          # Multi-tab endpoint creation & edit modal
│   │   └── index.css              # Custom monochrome theme & typography
│   └── package.json
├── wails.json                     # Wails v2 build configuration
└── Makefile                       # Automation & build targets
```

---

## 🛠️ Getting Started

### Prerequisites
- **Go 1.22+**
- **Node.js 18+ & npm**
- **Wails CLI v2**:
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

### System Dependencies

#### 🐧 Fedora / RHEL
```bash
sudo dnf install -y webkit2gtk4.1-devel gtk3-devel gcc
```

#### 🐧 Ubuntu / Debian
```bash
sudo apt-get update && sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev build-essential
```

---

## 🚀 Development & Build

### Run in Live Development Mode (Hot Reload)
```bash
wails dev -tags webkit2_41
```

### Build Production Binary
```bash
wails build -tags webkit2_41
```
The compiled executable will be located in `build/bin/vibeterm`.

---

## 📄 License
This project is open-source software licensed under the [MIT License](LICENSE).
