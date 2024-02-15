#!/usr/bin/env bash

curl -ski "https://localhost:3000/register?user=$1&token=$2"
