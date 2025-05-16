# Weather Subscription API Service

This service provides an API that allows users to subscribe to regular weather forecast updates for a chosen city. Users can manage their subscriptions via API endpoints, including confirmation via email.

*This project was created as a technical task for Genesis & KMA SOFTWARE ENGINEERING SCHOOL 5.0*

## Features

*   Get current weather for a specified city.
*   Subscribe to weather updates for a city (hourly or daily frequency).
*   Email confirmation for new subscriptions (**Note:** Currently, email content is logged to the console instead of being sent via a live email server).
*   Unsubscribe from weather updates.
*   Scheduled delivery of weather forecasts to confirmed subscribers (**Note:** Email content is logged to console).
*   Simple HTML page for user subscription.
*   Dockerized application for easy setup and deployment.

## API Endpoints

The API base path is `/api`.

| Method | Path                    | Description                               |
| :----- | :---------------------- | :---------------------------------------- |
| `GET`  | `/weather`              | Get current weather for a city.           |
| `POST` | `/subscribe`            | Subscribe to weather updates.             |
| `GET`  | `/confirm/{token}`      | Confirm email subscription.               |
| `GET`  | `/unsubscribe/{token}`  | Unsubscribe from weather updates.         |

You can see full swagger api requirements [here.](https://github.com/mykhailo-hrynko/se-school-5/blob/c05946703852b277e9d6dcb63ffd06fd1e06da5f/swagger.yaml)

## Running with Docker

1. Make sure ports 5432 and 8080 are available

2. Copy .env.example to .env and update the values

```
cp .env.example .env
```

3. Build and start services
```
docker compose up --build
```

4. The application will be accessible at http://localhost:8080. The PostgreSQL database will be accessible on host port 5432.

5. To stop the services
```
docker compose down
```

## Project structure

```
/weather-app
├── cmd/weatherapi_service/main.go  # Main application entry point
├── internal/
│   ├── api/                        # HTTP handlers, routing
│   ├── core/                       # Core domain models (Weather, Subscription)
│   ├── platform/                   # Concrete implementations (DB, email, scheduler, weather provider)
│   └── service/                    # Business logic services (subscription service)
├── migrations/                     # SQL migration files
├── web/                            # Static HTML/CSS/JS files for the frontend
│   └── index.html
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
└── .env
└── .env.example
```

## HTML Subscription Page
A simple HTML page is provided for subscribing to weather updates.
- If running the application (locally or via Docker), it's accessible at the root URL (e.g., http://localhost:8080/).
- The page allows users to enter their email, city, and desired update frequency.
- It also includes forms for manually confirming or unsubscribing using tokens (primarily for testing/demonstration).
- **To get confirm/unsubscribe token, check Docker logs for them after creating a subscription**

## Email Handling (Development Note)

Currently, this application **does not send actual emails** via an SMTP server or an external email service provider. For development and testing purposes, the email sending functionality is simulated by logging the intended email content (recipient, subject, body, and any links) to the application's console output.

**To "receive" an email (e.g., a confirmation or unsubscribe link):**
1.  Perform an action that would trigger an email (e.g., subscribe to updates).
2.  Check the console logs of the running application (either in your local terminal or via `docker compose logs app`).
3.  You will find log entries similar to:
    ```
    --- SENDING CONFIRMATION EMAIL ---
    To: user@example.com
    City: London
    Subject: Confirm your Weather Subscription for London
    Body: Please confirm your subscription by clicking this link: http://localhost:8080/api/confirm/your-confirmation-token
    --- END EMAIL ---
    ```
4.  You can then copy the relevant link (e.g., the confirmation link) from the log and paste it into your browser or use it with `curl` to proceed with the action (like confirming a subscription).

This approach was chosen to simplify setup for this technical task and avoid the need to configure external SMTP services or manage email credentials. The `internal/platform/email/service.go` file contains a `LogEmailService` which implements this logging behavior. The structure allows for a real SMTP service to be easily integrated in the future by creating a new implementation of the `email.Service` interface.
