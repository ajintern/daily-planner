# Daily Planner Application

A full-stack daily planner application developed using GoLang for the backend and Vanilla JavaScript for the frontend. The system utilizes Firebase Firestore as the primary database and follows a decoupled monorepo architecture.

## Project Structure

- **backend-monorepo/**: Contains the Go server implementation, API handlers, and Firebase integration logic.
- **frontend-monorepo/**: Contains the HTML, CSS, and JavaScript files responsible for the user interface and client-side interactions.

## Local Setup Instructions

### Backend Configuration

1. Navigate to the `backend-monorepo` directory.
2. Ensure Go is installed on the local machine.
3. Place the Firebase service account key file (`serviceAccountKey.json`) in the root of the backend directory. Note that this file is excluded from version control via `.gitignore`.
4. Install required dependencies by executing:
   ```bash
   go mod tidy
