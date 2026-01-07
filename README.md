# Glog

Glog is a minimalist, local-first journaling and knowledge base application built with [Wails](https://wails.io/), Go, and Svelte. It combines the performance of a native Go backend with the reactivity of a Svelte frontend, focusing on a distraction-free writing experience.

<p align="center">
  <img src="build/appicon.png" alt="Glog Icon" width="128" />
</p>

## Features

-   **Daily Journaling**: Automatically opens to today's entry, encouraging daily reflection.
-   **Block-Based Editor**: A rich text editor that treats paragraphs as blocks, similar to Notion.
-   **Bi-Directional Linking**: Connect your thoughts and documents using `[[WikiLink]]` syntax.
-   **Infinite Scroll**: Seamlessly browse through your journal history without pagination clicks.
-   **Scheduled Tasks**: Integrated task management within your journal entries.
-   **Local & Private**: All data is stored locally in a highly efficient BoltDB database (`glog.db`). No cloud required.
-   **Dark Mode**: A carefully crafted dark theme with accent colors for long writing sessions.

## Tech Stack

-   **Frontend**: Svelte, TypeScript, Vite
-   **Backend**: Go (Golang)
-   **Framework**: Wails v2
-   **Database**: BoltDB (bbolt)

## Getting Started

### Prerequisites

-   [Go](https://go.dev/) (v1.21 or newer)
-   [Node.js](https://nodejs.org/) (v16 or newer)

### Installation

1.  **Clone the repository**
    ```bash
    git clone https://github.com/yourusername/glog.git
    cd glog
    ```

2.  **Install project dependencies**
    Go dependencies (automatically handled by Wails/Go, but you can verify):
    ```bash
    go mod tidy
    ```
    Frontend dependencies:
    ```bash
    cd frontend
    npm install
    ```

### Live Development

To run the application in development mode with hot-reloading:

```bash
wails dev
```

This will start the application window and a Vite development server. You can also access the frontend in your browser at `http://localhost:34115` to use browser-based devtools.

### Building for Production

To create a standalone executable for your operating system:

```bash
wails build
```

The compiled binary will be located in the `build/bin/` directory.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

