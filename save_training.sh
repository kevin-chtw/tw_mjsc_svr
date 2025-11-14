#!/bin/bash
# 训练成果保存脚本

cd /root/pitaya/tw_mjsc_svr

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="training_backups/${TIMESTAMP}"

echo "========================================="
echo "   保存训练成果"
echo "   时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "========================================="
echo ""

# 1. 检查 Python 服务状态
echo "📡 1. 检查服务状态..."
if ! curl -s http://localhost:50051/health > /dev/null 2>&1; then
    echo "  ❌ Python AI 服务未运行，无法保存"
    exit 1
fi

HEALTH=$(curl -s http://localhost:50051/health)
TRAIN_COUNT=$(echo $HEALTH | python3 -c "import sys,json; print(json.load(sys.stdin)['train_count'])" 2>/dev/null)
BUFFER_SIZE=$(echo $HEALTH | python3 -c "import sys,json; print(json.load(sys.stdin)['buffer_size'])" 2>/dev/null)
EPSILON=$(echo $HEALTH | python3 -c "import sys,json; print(f\"{json.load(sys.stdin)['epsilon']:.4f}\")" 2>/dev/null)

echo "  ✅ 服务正常"
echo "     训练步数: $TRAIN_COUNT"
echo "     Buffer: $BUFFER_SIZE"
echo "     Epsilon: $EPSILON"
echo ""

# 2. 触发模型保存
echo "💾 2. 触发模型保存..."
curl -s -X POST -H "Content-Type: application/json" -d '{}' http://localhost:50051/save_model > /dev/null 2>&1
sleep 2

if [ -f "python_ai_service/mahjong_dqn.pth" ]; then
    MODEL_SIZE=$(ls -lh python_ai_service/mahjong_dqn.pth | awk '{print $5}')
    echo "  ✅ 模型已保存"
    echo "     文件: python_ai_service/mahjong_dqn.pth"
    echo "     大小: $MODEL_SIZE"
else
    echo "  ❌ 模型文件不存在"
    exit 1
fi
echo ""

# 3. 创建备份目录
echo "📦 3. 创建备份..."
mkdir -p "$BACKUP_DIR"

# 复制模型文件
cp python_ai_service/mahjong_dqn.pth "$BACKUP_DIR/"

# 复制日志文件
cp python_ai_service/ai_service.log "$BACKUP_DIR/"
cp trainer/logs/trainer-20251114.log "$BACKUP_DIR/"

# 创建元数据文件
cat > "$BACKUP_DIR/metadata.txt" << EOF
训练成果备份
===================
备份时间: $(date '+%Y-%m-%d %H:%M:%S')

训练状态:
- 训练步数: $TRAIN_COUNT
- Buffer 大小: $BUFFER_SIZE
- Epsilon: $EPSILON

游戏统计:
- 总游戏数: $(grep -c 'Game over' trainer/logs/trainer-20251114.log 2>/dev/null)
- Episodes: $(grep -c '📦 Episode' python_ai_service/ai_service.log 2>/dev/null)
- 胡牌次数: $(grep -c 'isHu=true' trainer/logs/trainer-20251114.log 2>/dev/null)
- 流局次数: $(grep -c 'liuju=true' trainer/logs/trainer-20251114.log 2>/dev/null)

训练效果:
- 平均 Loss: $(grep "🔥 Train" python_ai_service/ai_service.log | awk -F'loss=' '{print $2}' | awk -F',' '{sum+=$1; count++} END {printf "%.6f", sum/count}' 2>/dev/null)

文件清单:
- mahjong_dqn.pth (模型权重)
- ai_service.log (Python 服务日志)
- trainer-20251114.log (训练日志)
EOF

echo "  ✅ 备份创建成功"
echo "     目录: $BACKUP_DIR"
echo ""

# 4. 显示备份内容
echo "📋 4. 备份内容"
echo "----------------------------------------"
ls -lh "$BACKUP_DIR"
echo ""

# 5. 创建最新链接
ln -sf "$BACKUP_DIR" training_backups/latest
echo "  ✅ 已创建最新备份链接: training_backups/latest"
echo ""

# 6. 统计所有备份
BACKUP_COUNT=$(ls -1 training_backups/ 2>/dev/null | grep -E '^[0-9]{8}_[0-9]{6}$' | wc -l)
echo "📊 5. 备份统计"
echo "----------------------------------------"
echo "  总备份数: $BACKUP_COUNT"
echo "  备份位置: $(pwd)/training_backups/"
echo ""

# 7. 显示如何恢复
echo "💡 6. 如何恢复备份"
echo "----------------------------------------"
echo "  1. 停止训练服务:"
echo "     pkill -f 'python.*ai_service'"
echo ""
echo "  2. 恢复模型文件:"
echo "     cp $BACKUP_DIR/mahjong_dqn.pth python_ai_service/"
echo ""
echo "  3. 重启服务:"
echo "     cd python_ai_service && python3 -u ai_service.py >> ai_service.log 2>&1 &"
echo ""

echo "========================================="
echo "  ✅ 保存完成"
echo "========================================="

