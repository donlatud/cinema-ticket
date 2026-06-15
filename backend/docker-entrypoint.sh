#!/bin/sh
set -e

CREDENTIALS_PATH="${FIREBASE_CREDENTIALS:-/app/firebase-key.json}"

if [ ! -e "$CREDENTIALS_PATH" ]; then
  echo "ERROR: Firebase credentials not found at $CREDENTIALS_PATH"
  echo ""
  echo "Before running docker compose:"
  echo "  1. Download Service Account JSON from Firebase Console"
  echo "  2. Save it as: backend/firebase-key.json"
  echo ""
  echo "On Windows: if the file is missing, Docker may create a FOLDER named"
  echo "firebase-key.json instead. Delete that folder and add the real JSON file."
  exit 1
fi

if [ -d "$CREDENTIALS_PATH" ]; then
  echo "ERROR: $CREDENTIALS_PATH is a directory, not a file."
  echo "Delete the folder and place your Firebase service account JSON file there."
  exit 1
fi

if ! head -c 1 "$CREDENTIALS_PATH" | grep -q '{'; then
  echo "ERROR: $CREDENTIALS_PATH does not look like valid JSON."
  exit 1
fi

exec /app/server
