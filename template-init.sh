#!/bin/bash

set -euo pipefail

# Get repo name from argument or git
REPO_NAME="${1:-$(basename -s .git "$(git config --get remote.origin.url)")}"
TYPE_NAME="$(echo "$REPO_NAME" | sed -r 's/(^|-)([a-z])/\U\2/g')API"

echo "🔧 Renaming template to:"
echo " - Repo/module: $REPO_NAME"
echo " - GraphQL type: $TYPE_NAME"

# Replace in files
find . -type f \( -name '*.go' -o -name '*.graphql' -o -name '*.gql' -o -name '*.mod' -o -name '*.yaml' -o -name '*.yml' \) \
  -not -path "./.git/*" \
  -exec sed -i "s/golang-template/${REPO_NAME}/g" {} +

find . -type f \( -name '*.go' -o -name '*.graphql' -o -name '*.gql' \) \
  -not -path "./.git/*" \
  -exec sed -i "s/golangTemplateAPI/${TYPE_NAME}/g" {} +

# Rename folders
find . -type d -name 'golang-template' | while read -r dir; do
  newdir=$(echo "$dir" | sed "s/golang-template/${REPO_NAME}/")
  mv "$dir" "$newdir"
done

# Tidy modules & run gqlgen
echo "🧹 Running go mod tidy..."
go mod tidy

echo "📦 Running gqlgen generate..."
go run github.com/99designs/gqlgen generate

# Commit and push (skip workflows)
git config user.name "github-actions"
git config user.email "github-actions@github.com"

git add . ':!**/.github/workflows/*'

if git diff --cached --quiet; then
  echo "✅ No changes to commit."
else
  git commit -m "chore: rename template and generate gql files"
  git push origin main
  echo "🚀 Changes pushed."
fi
