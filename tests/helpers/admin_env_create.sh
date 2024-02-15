#!/usr/bin/env bash

# required fields:
#   action: env-create: Add a new environment
#   name: The subject of the action
#   token: Admin token

curl -ski "https://localhost:3000/admin?action=env-create&name=$1&token=$2"


