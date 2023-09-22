# GoShort

GoShort is a URL shortening service built with Go. It provides a simple and efficient way to shorten long URLs and retrieve the original URL using the shortened version.

## Features

- **URL Shortening**: Convert long URLs into shortened versions.
- **URL Retrieval**: Retrieve the original URL using the shortened version.
- **Visit Counter**: Track the number of visits for each shortened URL.
- **Caching**: Utilizes Redis for caching to improve performance.
- **MongoDB Integration**: Uses MongoDB for persistent storage of URLs.
- **Docker Support**: Easily deployable using Docker.

## Getting Started

### Prerequisites

- Go (version 1.20 or higher)
- Docker & Docker Compose
- MongoDB
- Redis

### Installation

1. Clone the repository:

git clone https://github.com/hajbabaeim/goshort.git

2. Navigate to the project directory:

cd goshort

Copy code

3. Build the Docker image:

docker-compose build

Copy code

4. Start the services using Docker Compose:

docker-compose up

The application will be accessible at the specified port (default: 8080).

## API Endpoints

- **Shorten URL**: `POST /short`
- Request Body: `{ "url": "YOUR_LONG_URL" }`
- Response: JSON object with the shortened URL.

- **Retrieve Original URL**: `GET /:shortUrl`
- Response: Redirects to the original URL.

- **Get All Shortened URLs**: `GET /`
- Response: JSON array of all shortened URLs.

## Configuration

The application can be configured using environment variables. Refer to the `.env` file for available configuration options.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## License

This project is open-source and available under the MIT License.
