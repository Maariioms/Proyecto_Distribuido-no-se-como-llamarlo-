#!/usr/bin/env bash
set -e

echo "== XSIAM-sito: script de automatización =="

if ! command -v docker &> /dev/null; then
	echo "Docker ha morido"
	exit 1
fi

if ! docker info > /dev/null 2>&1; then
	echo "Docker esta en el lobby pero no en la partida"
	exit 1
fi

echo "Docker ahorita levanta a todos"
docker compose up --build

echo "Frontend :3000, middleware en :9000"
