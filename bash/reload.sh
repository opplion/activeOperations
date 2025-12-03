#!/bin/sh
set -eu

if [ -d "website" ]; then
    echo "🗑️  检测到 website 目录，正在删除以确保重新克隆..."
    rm -rf website
fi

echo "📥 开始 clone website 仓库（仅 content/zh-cn/docs）..."

git clone --filter=blob:none --sparse -b main https://github.com/kubernetes/website.git website

cd website

git sparse-checkout init --cone
git sparse-checkout set content/zh-cn/docs/concepts

echo "✅ website 克隆完成！"
