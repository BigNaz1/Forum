#!/bin/bash

# Stop and remove any existing container with the same name
docker stop reboot-forums 2>/dev/null
docker rm reboot-forums 2>/dev/null

# Build the Docker image
echo "Building Docker image..."
docker build -t reboot-forums:latest .

# Run the Docker container
echo "Running Docker container..."
docker run -d --name reboot-forums -p 8080:8080 reboot-forums:latest

# Check if the container is running
if [ "$(docker ps -q -f name=reboot-forums)" ]; then
    echo "RebootForums is now running on http://localhost:8080"
else
    echo "Failed to start RebootForums container. Check Docker logs for more information."
fi

# Display the logs
echo "Displaying container logs:"
docker logs reboot-forums   