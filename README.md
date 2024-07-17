# Jambda - A Self-Hosted Serverless Framework

## Overview
Jambda is a custom, self-hosted serverless framework designed to execute function binaries in response to specific triggers. It is similar to AWS Lambda but tailored for personal and home use. This project serves both as an educational project into the workings of serverless architectures and a practical solution for automating and managing home-based tasks such as scheduled backups, API serving, and event-driven processing.

## Features
- Golang & Java support Binary/Jar support
- Triggers via HTTP and CRON events (WIP)
- Isolation through Docker
- Multi-instance log streaming (WIP)
- Frontend for configuration, and viewing logs (WIP)
- Automatic scaling (WIP - currently only one instance that scales down to 0)
- **Hot/Cold Starts**: Containers automatically shutdown after prolonged periods of no use. A new request will create a new instance of the application function, with subsequent requets having much better performance.
- **Scaling**: Monitors metrics such as requests per second/minute to scale down functions when not in use.
- **HTTP Trigger System:** Functions can be triggered via HTTP requests, making the system versatile and easy to integrate with existing home networks or internet-based services.
- **Modular Function Design:** Each function is isolated, allowing for targeted updates and maintenance without affecting the entire system.
- **Scalability and Concurrency:** Built using Golang’s robust concurrency model, allowing multiple functions to be executed simultaneously without performance bottlenecks.

## Future features
- On the fly configuration updates
- Kubernetes integration for managing multiple pods, allowing for better load balancing, minimal cold starts and zero downtime updates.


## Technical Approach
- **HTTP Server in Go:** The main system of Jambda, handling incoming requests and routing them to specific functions, based on configuration.
- **Docker for Isolation:** Utilize Docker to run functions in isolated containers, mimicking the separation found in professional serverless environments.
- **Dynamic Function Registration:** Functions can be added or updated using the configuration frontend, with the system dynamically routing requests based on the function id provided.
- **Logging and Monitoring:** (WIP) Logging for all function executions to monitor performance and troubleshoot issues, via the dashboard.

## Running Jambda

### Installation Prerequisites
- Go 1.22+ installed
- Docker

### Setup
*TODO*

### Function configuration rules
*TODO*
