# test1
# National Training API

## Overview

This is a robust, secure, and production-ready RESTful API built with Go for managing a **National In-Service Training Database**. It provides a comprehensive set of endpoints for handling officers, courses, sessions, attendance, and user management, built with a focus on security, concurrency, and best practices.

## Features

-   **Full CRUD Operations** for core resources:
    -   Officers
    -   Courses
    -   Sessions
    -   Facilitators
    -   Attendance Records
-   **Complete User Authentication & Authorization Flow**:
    -   Secure user registration with password hashing (`bcrypt`).
    -   User account activation via email using secure, single-use tokens.
    -   Bearer Token authentication for accessing protected endpoints.
    -   Secure password reset flow via email.
-   **Background Worker Jobs**: Emails are sent asynchronously in background goroutines to ensure fast API response times.
-   **Secure by Design**:
    -   Configurable Cross-Origin Resource Sharing (CORS) to protect against browser-based attacks.
    -   Rate limiting to prevent abuse.
    -   Graceful shutdown to ensure all tasks complete before the server exits.
-   **Database Management**:
    -   Uses PostgreSQL for data persistence.
    -   Includes a complete database migration system using `golang-migrate`.
-   **Configuration Driven**:
    -   Configuration managed via command-line flags and environment variables (`.env` file).

## Prerequisites

Before you begin, ensure you have the following installed on your system:
-   **Go** (version 1.18 or newer)
-   **PostgreSQL** (version 12 or newer)
-   **Make**
-   **[golang-migrate/migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)** for database migrations.

## Getting Started

Follow these steps to get the API up and running on your local machine.

### 1. Clone the Repository

```bash
git clone https://github.com/amari03/test1.git
cd test1
```

### 2. Set Up the Database

You need to create a PostgreSQL database and a dedicated user for the application.
```SQL
-- Connect to psql as a superuser
CREATE ROLE youruser WITH LOGIN PASSWORD 'yourpassword';
CREATE DATABASE yourdb OWNER youruser;
```

DO NOT FORGET TO UPDATE PERMISSIONS ALSO!

This is done as the superuser:

```SQL
postgres=# ALTER DATABASE comments OWNER TO comments;
postgres=# GRANT CREATE ON DATABASE comments TO comments;
```

### 3. Configure Your Environment

The application uses an .envrc file to manage secrets and environment-specific configuration.

Create a file named .envrc in the root of the project and populate it with your credentials.

.envrc:
```Sh
# PostgreSQL Database Connection String
export COMMENTS_DB_DSN="postgres://youruser:yourpassword@localhost/yourdb?sslmode=disable"

# Mailtrap SMTP Credentials for Email
export SMTP_HOST="sandbox.smtp.mailtrap.io"
export SMTP_PORT="2525"
export SMTP_USERNAME="your-mailtrap-username"
export SMTP_PASSWORD="your-mailtrap-password"
export SMTP_SENDER="Your App Name <no-reply@yourapp.com>"
```

### 4. Remember to set up a Mailtrap account

The purpose of Mailtrap will provide us with a testing inbox that we can send our emails to. This way we can see check that our app would be able to send real emails correctly to real users when the app goes into production.

Your credentials add them to your .envrc file.

### 5. Load Environment Variables

Load the variables from your .env file into your current shell session. You'll need to do this every time you open a new terminal.

```Bash
source .envrc
```

### 6. Install Go Dependencies

Download the necessary GO modules

```Bash
go mod tidy
```

### 7. Run Database Migrations

Apply the databse schema to your newly created database

```Bash
make db/migrations/up
```

### 8. Run Application

You are now ready to start the API server!

```Bash
make run/api
```

 -The API server should now be running on http://localhost:4000.  
 -Curl commands are also available to use for testing

## License
Project created by Addie Vasquez and Mickali Garbutt 2025