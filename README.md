# Video Streaming

A robust video microservice built with Go that handles video upload, processing, and streaming using HLS and gRPC. This project provides a complete solution for uploading large video files, processing them, and streaming them to a web player.

## Features

- **Large File Handling**: Efficiently handles gigabyte-sized video files.
- **Video Processing**: Uses FFmpeg for video encoding and processing.
- **Streaming**: Implements HLS (HTTP Live Streaming) for adaptive streaming.
- **gRPC Integration**: Communicates with services using gRPC for efficient, high-performance data transfer.
- **Security**: Includes token-based authentication for secure access to video segments.
- **CORS Handling**: Configured to handle cross-origin requests.

## Architecture

- **Go**: The main programming language used for backend development.
- **FFmpeg**: For video processing and encoding.
- **HLS.js**: JavaScript library for playing HLS streams in the browser.
- **gRPC**: For inter-service communication.
- **gofr**: Microservice framework used in Go for building the microservices.
- **Echo**: HTTP framework used in Go for building the API.
