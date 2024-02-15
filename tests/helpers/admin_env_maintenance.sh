#!/usr/bin/env bash

# required fields:
#   action: env-maintenance: Setup an env for maintenance
#   name: The subject of the action
#   token: Admin token

curl -ski "https://localhost:3000/admin?action=env-maintenance&name=$1&token=$2"


