#!/bin/bash

# Post request to the server
curl \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"key1":"value1", "key2":"value2"}' \
    http://localhost:8080