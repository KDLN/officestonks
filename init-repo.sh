#!/bin/bash

# This script initializes the Git repository and makes the first commit

# Check if git is installed
if ! command -v git &> /dev/null; then
    echo "Git is not installed. Please install Git first."
    exit 1
fi

# Initialize Git repository
git init

# Add all files
git add .

# Make initial commit
git commit -m "Initial commit: Office Stonks MVP project structure"

# Instructions for linking to GitHub
echo ""
echo "Repository initialized with initial commit."
echo ""
echo "To link to a GitHub repository, run the following commands:"
echo ""
echo "  git remote add origin https://github.com/yourusername/officestonks.git"
echo "  git branch -M main"
echo "  git push -u origin main"
echo ""
echo "Replace 'yourusername' with your GitHub username."