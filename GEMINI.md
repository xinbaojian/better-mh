# GEMINI Project: better-mh

## Project Overview

This project, `better-mh`, is a desktop application built with the [Wails](https://wails.io/) framework. It features a Go backend and a Vue.js frontend. The application is designed to assist with a game, likely a mobile game running on an Android device, by providing automation and assistance features. The use of the `gocv` library in the backend suggests that it performs computer vision tasks, such as image recognition, to interact with the game. The frontend is built with Vue.js and the Element Plus UI library.

## Building and Running

### Prerequisites

- Go
- Node.js and npm
- Wails CLI

### Development

To run the application in live development mode, use the following command in the project root directory:

```bash
wails dev
```

This will start a development server for the frontend with hot-reloading and make the Go backend methods available to the frontend.

### Production

To build a production version of the application, use the following command:

```bash
wails build
```

This will create a distributable, production-ready package.

## Development Conventions

### Backend

The backend is written in Go. The main entry point is `main.go`, which initializes the Wails application and its various components, including:

- `App`: The main application logic.
- `Adb`: Handles communication with Android devices via ADB.
- `Game`: Contains the game-specific automation logic.
- `Message`: Manages messaging between the backend and frontend.
- `Log`: Handles logging.

### Frontend

The frontend is a Vue.js application located in the `frontend` directory. Key technologies include:

- **Vite:** The build tool for the frontend.
- **Vue.js:** The core JavaScript framework.
- **Element Plus:** A UI component library for Vue.js.

The frontend source code is in `frontend/src`. The main application component is `frontend/src/App.vue`.

### Communication

The Go backend and Vue.js frontend communicate via bindings that are defined in `main.go`. The Go methods are exposed to the frontend and can be called directly from the JavaScript code.
