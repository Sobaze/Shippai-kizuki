# Shippai Kizuki 失敗気づき

A personal Japanese learning and analytics application built around:

**Mistake → Insight → Learning → Improvement**

The project starts with a React + TypeScript frontend and a Go backend. Vocabulary
and kanji mistake analysis, SQL persistence, and Anki integration will be explored
in small iterations.

## Tech stack

- **Frontend:** React + TypeScript, with Vite for development and builds
- **Backend:** Go
- **Database:** SQL-based; selection planned for a later iteration

## Installation and setup

### 1. Install prerequisites

Install these tools on your computer:

| Tool | Version used during setup | Download |
| --- | --- | --- |
| Git | Any recent version | [git-scm.com](https://git-scm.com/downloads) |
| Node.js (includes npm) | 24 | [nodejs.org](https://nodejs.org/) |
| Go | 1.27 | [go.dev](https://go.dev/dl/) |

On Windows, you can install them with Windows Package Manager:

```powershell
winget install --id Git.Git --exact --source winget
winget install --id OpenJS.NodeJS.LTS --exact --source winget
winget install --id GoLang.Go --exact --source winget
```

Skip tools you already have installed. After installation, reopen your terminal
(and restart VS Code if using its integrated terminal) to pick up PATH changes.

Check that the tools are available:

```powershell
git --version
node --version
npm --version
go version
```

### 2. Clone the repository

```powershell
git clone https://github.com/Sobaze/Shippai-kizuki.git
cd Shippai-kizuki
```

If you already have the repository locally, open a terminal at its root instead.
Both `frontend/` and `backend/` belong to this single Git repository.

### 3. Install frontend dependencies

From the repository root:

```powershell
cd frontend
npm ci
```

`npm ci` installs the dependency versions recorded in `package-lock.json`.
Use it for a fresh checkout. When intentionally adding or updating dependencies,
use `npm install` instead.

The backend currently uses only Go's standard library, so it requires no
third-party dependency installation. Its Go module is already initialized.

## Running locally

### Frontend

From the repository root:

```powershell
cd frontend
npm run dev
```

Open the local URL printed by Vite (normally **http://localhost:5173**).
The frontend currently shows the React/Vite starter screen. Keep this terminal
running while using the frontend; press **Ctrl+C** to stop the development server.

### Backend

In a separate terminal, from the repository root:

```powershell
cd backend
go run .
```

Expected output:

```text
Shippai Kizuki backend is ready.
```

The backend currently prints this message and exits. An HTTP server and the
frontend-to-backend connection will be added in a future iteration.

## Development checks

Inside `frontend/`:

```powershell
npm run build
npm run lint
```

- `npm run build` checks TypeScript and produces a production build in `frontend/dist/`.
- `npm run lint` checks the frontend code for potential issues.

Inside `backend/`:

```powershell
go vet ./...
```

`go vet` checks Go code for common mistakes.

## Project structure

```text
Shippai-kizuki/
├── backend/
│   ├── go.mod          # Go module identity and Go version
│   └── main.go         # Backend executable entry point
├── frontend/
│   ├── public/         # Static assets
│   ├── src/            # React components, styles, and application entry point
│   ├── package.json    # Frontend dependencies and npm commands
│   └── package-lock.json
├── .gitignore
└── README.md
```

The Go module path matches this repository's GitHub location; running it locally
does not require publishing it.

## Troubleshooting

- **`go`, `node`, or `npm` is not recognized:** reopen your terminal after installing
  the tool. If using VS Code, restart it too. Verify the installation using the
  version commands above.
- **Frontend dependencies are missing:** run `npm ci` inside `frontend/`.
- **Port 5173 is already in use:** use the alternative local URL printed by Vite.
- **Commands cannot find `package.json` or `go.mod`:** run npm commands inside
  `frontend/` and Go commands inside `backend/`.
