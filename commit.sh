#!/bin/bash

echo "Commiting Auth Service"

if [ $# -lt 2 ]; then
    echo "Usage: $0 <commit_message> <branch_name>"
    exit 1
fi
 
git checkout -b "$2"

git add .

git commit -m "$1"

git push -u origin "$2"