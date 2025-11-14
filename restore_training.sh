#!/bin/bash
# 训练成果恢复脚本

cd /root/pitaya/tw_mjsc_svr

if [ -z "$1" ]; then
    echo "========================================="
    echo "   可用的训练备份"
    echo "========================================="
    echo ""
    
    if [ ! -d "training_backups" ]; then
        echo "  ❌ 没有找到备份目录"
        exit 1
    fi
    
    BACKUPS=$(ls -1 training_backups/ 2>/dev/null | grep -E '^[0-9]{8}_[0-9]{6}$' | sort -r)
    
    if [ -z "$BACKUPS" ]; then
        echo "  ❌ 没有找到任何备份"
        exit 1
    fi
    
    echo "序号 | 备份时间           | 训练步数 | Buffer | 游戏数"
    echo "-----+--------------------+----------+--------+-------"
    
    INDEX=1
    for BACKUP in $BACKUPS; do
        if [ -f "training_backups/$BACKUP/metadata.txt" ]; then
            TRAIN_COUNT=$(grep "训练步数:" training_backups/$BACKUP/metadata.txt | awk -F': ' '{print $2}')
            BUFFER=$(grep "Buffer 大小:" training_backups/$BACKUP/metadata.txt | awk -F': ' '{print $2}')
            GAMES=$(grep "总游戏数:" training_backups/$BACKUP/metadata.txt | awk -F': ' '{print $2}')
            
            YEAR=${BACKUP:0:4}
            MONTH=${BACKUP:4:2}
            DAY=${BACKUP:6:2}
            HOUR=${BACKUP:9:2}
            MIN=${BACKUP:11:2}
            SEC=${BACKUP:13:2}
            
            printf "%4d | %s-%s-%s %s:%s:%s | %8s | %6s | %5s\n" \
                $INDEX $YEAR $MONTH $DAY $HOUR $MIN $SEC \
                "$TRAIN_COUNT" "$BUFFER" "$GAMES"
        fi
        INDEX=$((INDEX + 1))
    done
    
    echo ""
    echo "💡 使用方法:"
    echo "   ./restore_training.sh <备份目录名>"
    echo "   或使用 'latest' 恢复最新备份:"
    echo "   ./restore_training.sh latest"
    echo ""
    exit 0
fi

BACKUP_NAME="$1"

if [ "$BACKUP_NAME" = "latest" ]; then
    BACKUP_DIR="training_backups/$(ls -1t training_backups/ | grep -E '^[0-9]{8}_[0-9]{6}$' | head -1)"
else
    BACKUP_DIR="training_backups/$BACKUP_NAME"
fi

if [ ! -d "$BACKUP_DIR" ]; then
    echo "❌ 备份目录不存在: $BACKUP_DIR"
    exit 1
fi

echo "========================================="
echo "   恢复训练成果"
echo "========================================="
echo ""

# 显示备份信息
if [ -f "$BACKUP_DIR/metadata.txt" ]; then
    echo "📄 备份信息:"
    cat "$BACKUP_DIR/metadata.txt"
    echo ""
fi

read -p "确认恢复此备份? (y/n): " confirm
if [ "$confirm" != "y" ]; then
    echo "❌ 取消恢复"
    exit 0
fi

echo ""
echo "🔄 开始恢复..."
echo ""

# 1. 停止服务
echo "1. 停止训练服务..."
pkill -f "python.*ai_service.py"
pkill -f "./trainer"
sleep 3
echo "  ✅ 服务已停止"
echo ""

# 2. 恢复模型文件
echo "2. 恢复模型文件..."
if [ -f "$BACKUP_DIR/mahjong_dqn.pth" ]; then
    cp "$BACKUP_DIR/mahjong_dqn.pth" python_ai_service/
    echo "  ✅ 模型文件已恢复"
    ls -lh python_ai_service/mahjong_dqn.pth
else
    echo "  ❌ 备份中没有模型文件"
    exit 1
fi
echo ""

# 3. 重启 Python 服务
echo "3. 重启 Python AI 服务..."
cd python_ai_service
python3 -u ai_service.py >> ai_service.log 2>&1 &
PYTHON_PID=$!
echo "  Python PID: $PYTHON_PID"
sleep 5

if curl -s http://localhost:50051/health > /dev/null 2>&1; then
    echo "  ✅ Python 服务已启动"
    curl -s http://localhost:50051/health | python3 -m json.tool
else
    echo "  ❌ Python 服务启动失败"
    exit 1
fi
echo ""

# 4. 重启 Trainer
echo "4. 重启 Go Trainer..."
cd /root/pitaya/tw_mjsc_svr/trainer
./trainer > logs/trainer-restored.log 2>&1 &
TRAINER_PID=$!
echo "  Trainer PID: $TRAINER_PID"
echo ""

sleep 5

echo "========================================="
echo "  ✅ 恢复完成"
echo "========================================="
echo ""
echo "💡 训练已恢复并继续运行"
echo "   使用 ./watch_training.sh 监控训练状态"

