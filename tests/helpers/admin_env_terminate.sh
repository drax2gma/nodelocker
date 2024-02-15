#!/usr/bin/env bash

# required fields:
#   action: env-terminate: Lock an env indefinately
#   name: The subject of the action
#   token: Admin token

curl -ski "https://localhost:3000/admin?action=env-terminate&name=$1&token=$2"


