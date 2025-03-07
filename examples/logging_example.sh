#!/bin/bash
# Example script demonstrating the Datadog CLI with logging

# Build the CLI
echo "Building Datadog CLI..."
go build -o dd ./cmd/dd

# Check if API keys are set
if [ -z "$DD_API_KEY" ] || [ -z "$DD_APP_KEY" ]; then
    echo "Error: DD_API_KEY and DD_APP_KEY environment variables must be set."
    echo "Please set them with:"
    echo "export DD_API_KEY=your_api_key"
    echo "export DD_APP_KEY=your_application_key"
    exit 1
fi

# Set environment for logs
export DD_ENV="example"

echo "Running commands with logging enabled..."

# Run a series of commands to demonstrate logging
echo -e "\n1. List hosts with debug logging enabled:"
./dd --debug hosts list --limit 5

echo -e "\n2. List monitors with custom environment tag:"
./dd --env demo monitors list --limit 5

echo -e "\n3. List tags with JSON output:"
./dd --output json tags list

echo -e "\nAll commands have been executed with logging to Datadog."
echo "You can view these logs in your Datadog account by searching for:"
echo "  service:datadog-cli env:example"
echo "  service:datadog-cli env:demo" 