#!/usr/bin/env bash
# 服务器部署脚本
# 用法: ./deploy/deploy-server.sh [test|prod]

set -euo pipefail

ENV="${1:-}"
if [[ "$ENV" != "test" && "$ENV" != "prod" ]]; then
    echo "用法: $0 [test|prod]"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

IMAGE_NAME="sub2api:${ENV}"
CONTAINER_NAME="sub2api-${ENV}"
CONFIG_FILE="/etc/sub2api/${ENV}.yaml"
NETWORK="deploy_sub2api-network"
LOG_DIR="/data/logs/sub2api/${ENV}"

if [[ "$ENV" == "test" ]]; then
    HOST_PORT="127.0.0.1:8081"
else
    HOST_PORT="127.0.0.1:8080"
fi

mkdir -p "$LOG_DIR"

echo "==> [1/4] 拉取最新代码"
cd "$REPO_ROOT"
git pull

if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "错误: 配置文件不存在: ${CONFIG_FILE}"
    exit 1
fi

echo "==> [2/4] 构建镜像 ${IMAGE_NAME}"
docker build \
    -t "$IMAGE_NAME" \
    -f "${REPO_ROOT}/Dockerfile" \
    "${REPO_ROOT}"

echo "==> [3/4] 备份日志并移除旧容器"
if docker inspect "$CONTAINER_NAME" &>/dev/null; then
    docker logs "$CONTAINER_NAME" > "${LOG_DIR}/deploy-$(date +%Y%m%d_%H%M%S).log" 2>&1 || true
fi
docker stop "$CONTAINER_NAME" 2>/dev/null || true
docker rm "$CONTAINER_NAME" 2>/dev/null || true

echo "==> [4/4] 启动新容器"
docker run -d \
    --name "$CONTAINER_NAME" \
    --restart unless-stopped \
    --network "$NETWORK" \
    --ulimit nofile=100000:100000 \
    --log-driver json-file \
    --log-opt max-size=50m \
    --log-opt max-file=5 \
    -v "${CONFIG_FILE}:/app/data/config.yaml:ro" \
    -p "${HOST_PORT}:8080" \
    "$IMAGE_NAME"

echo "==> 部署完成: ${CONTAINER_NAME} -> ${HOST_PORT}"
