#!/bin/bash

# PentAGI 比赛模式重启脚本
# 修改完 docker-compose.yml 后运行此脚本

echo "========================================="
echo "PentAGI 比赛模式重启"
echo "========================================="

cd ~/Desktop/pentAGI || cd /Desktop/pentAGI

echo ""
echo "1. 停止所有容器..."
docker compose down

echo ""
echo "2. 验证 .env 配置..."
echo "COMPETITION配置："
cat .env | grep COMPETITION

echo ""
echo "3. 重新启动容器..."
docker compose up -d

echo ""
echo "4. 等待服务启动（10秒）..."
sleep 10

echo ""
echo "5. 验证环境变量..."
docker compose exec pentagi printenv | grep COMPETITION

echo ""
echo "6. 检查服务启动..."
docker compose logs pentagi | grep -i "competition service started"

echo ""
echo "========================================="
echo "如果看到 'Competition service started'，说明成功！"
echo "========================================="

echo ""
echo "实时监控日志："
echo "docker compose logs -f pentagi | grep -i competition"

