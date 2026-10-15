#!/usr/bin/env bash
sleep 300
while true; do
  sleep 120
  if ! pgrep -f 'vscode.*bootstrap-fork.*extensionHost' > /dev/null; then
    sudo kill -15 1
  fi
done