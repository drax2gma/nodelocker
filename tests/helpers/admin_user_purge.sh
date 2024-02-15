#!/usr/bin/env bash

# required fields:
#   action: user-purge: Purge a user which probably forgot their password
#   name: The subject of the action
#   token: Admin token

curl -ski "https://localhost:3000/admin?action=user-purge&name=$1&token=$2"


