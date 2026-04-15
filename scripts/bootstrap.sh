#!/usr/bin/env bash
set -euo pipefail

echo "Bootstrap: checking required tools"

need_install=()

command -v docker >/dev/null 2>&1 || need_install+=(docker.io)
command -v docker-compose >/dev/null 2>&1 || need_install+=(docker-compose)
command -v go >/dev/null 2>&1 || need_install+=(golang-go)
command -v node >/dev/null 2>&1 || need_install+=(nodejs)
command -v npm >/dev/null 2>&1 || need_install+=(npm)
command -v pnpm >/dev/null 2>&1 || need_install+=(pnpm)

if [ ${#need_install[@]} -eq 0 ]; then
  echo "All required tools are already installed."
  exit 0
fi

echo "About to install: ${need_install[*]}"
echo "This script will use sudo to install packages via apt. Continue? (y/N)"
read -r ans
if [ "${ans}" != "y" ] && [ "${ans}" != "Y" ]; then
  echo "Aborting. Run the bootstrap script manually when ready."
  exit 1
fi

sudo apt-get update
sudo apt-get install -y ${need_install[*]} curl unzip

# Install pnpm if missing
if ! command -v pnpm >/dev/null 2>&1; then
  sudo npm install -g pnpm
fi

echo "Bootstrap complete. You may now run 'make up' to build and start services."
