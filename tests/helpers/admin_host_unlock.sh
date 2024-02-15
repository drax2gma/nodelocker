#!/usr/bin/env bash

# required fields:
#   action: host-unlock: Unlock a stuck, locked host#   name: the subject of the action
#   name: The subject of the action
#   token: Admin token

curl -ski "https://localhost:3000/admin?action=host-unlock&name=$1&token=$2"


