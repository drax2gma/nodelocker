#!/usr/bin/env bash

curl -ski "https://localhost:3000/unlock?type=env&name=$1&user=$2&token=$3"
