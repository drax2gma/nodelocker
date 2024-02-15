#!/usr/bin/env bash

# required fields:
#   action: env-unlock: Unlock an env from maintenance or terminate state
#   name: The subject of the action
#   token: Admin token

curl -ski "https://localhost:3000/admin?action=env-unlock&name=$1&token=$2"


