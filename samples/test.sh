#!/bin/bash

# A simple bash script demo

APP_NAME="MyShellApp"
VERSION="1.0.0"

echo "Starting $APP_NAME version $VERSION..."

# Check if a directory exists
if [ -d "/tmp" ]; then
    echo "/tmp directory exists."
else
    echo "/tmp directory does not exist, creating it..."
    mkdir /tmp
fi

# Loop through arguments
echo "Arguments provided:"
for arg in "$@"; do
    echo "- $arg"
done

# Function definition
greet() {
    local name=$1
    echo "Hello, $name!"
}

# Call function
greet "User"

# While loop
count=0
while [ $count -lt 5 ]; do
    echo "Count: $count"
    ((count++))
done

# Case statement
case "$1" in
    start)
        echo "Starting service..."
        ;;
    stop)
        echo "Stopping service..."
        ;;
    restart)
        echo "Restarting service..."
        ;;
    *)
        echo "Usage: $0 {start|stop|restart}"
        exit 1
        ;;
esac

echo "Done."
