# ScaleBit Frontend Dashboard

## 1. Project Overview

This is the official frontend dashboard for the ScaleBit Microservices Platform. It provides a user interface for interacting with the platform's services, monitoring key metrics, and managing resources. The dashboard is built with React and Material-UI.

## 2. Getting Started

Follow these steps to run the frontend dashboard locally for development.

### Prerequisites

- Node.js (version 16 or later)
- npm or yarn

### Local Development Setup

1. **Navigate to the Dashboard Directory**:
   ```sh
   cd frontend-dashboard
   ```

2. **Install Dependencies**:
   ```sh
   npm install
   ```
   or
   ```sh
   yarn install
   ```

3. **Run the Development Server**:
   ```sh
   npm start
   ```
   This will start the application in development mode and open it in your default browser at `http://localhost:3000`. The page will automatically reload if you make edits.

## 3. Available Scripts

In the project directory, you can run the following scripts:

- **`npm start`**: Runs the app in development mode.
- **`npm test`**: Launches the test runner in interactive watch mode.
- **`npm run build`**: Builds the app for production to the `build` folder.
- **`npm run eject`**: Removes the single dependency configuration and copies all configuration files and transitive dependencies into your project. **Note: this is a one-way operation.**

## 4. Folder Structure

The `src` directory contains the main application code:

- **`api/`**: Contains the Axios instance and functions for making API calls to the backend services.
- **`pages/`**: Contains the main page components of the application (e.g., `Dashboard.js`, `Login.js`).
- **`App.js`**: The root component of the application, which handles routing.
- **`index.js`**: The entry point of the application.

## 5. API Integration

The dashboard communicates with the ScaleBit backend services through an API gateway. The base URL for the API is configured in `src/api/axios.js`. When running locally, you may need to update this URL to point to your local or staging environment's API gateway.

## 6. Contribution Guidelines

We welcome contributions to the frontend dashboard. To contribute, please follow these guidelines:

- **Fork the repository** and create a new branch for your feature or bug fix.
- **Follow the existing code style** and component structure.
- **Write clear and professional commit messages**.
- **Ensure the application builds successfully** before submitting a pull request.
- **Submit a pull request** for review.

## 7. License

This project is licensed under the MIT License.
