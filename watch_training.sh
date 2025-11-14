#!/bin/bash
# 训练监控脚本

cd /root/pitaya/tw_mjsc_svr

while true; do
    clear
    echo "========================================="
    echo "   麻将 AI 训练监控"
    echo "   $(date '+%Y-%m-%d %H:%M:%S')"
    echo "========================================="
    echo ""
    
    # 检查进程
    echo "📊 进程状态:"
    PYTHON_PID=$(ps aux | grep "python.*ai_service.py" | grep -v grep | awk '{print $2}')
    TRAINER_PID=$(ps aux | grep "./trainer" | grep -v grep | awk '{print $2}')
    
    if [ -n "$PYTHON_PID" ]; then
        echo "  ✅ Python AI Service: PID $PYTHON_PID"
    else
        echo "  ❌ Python AI Service: 未运行"
        echo "     正在重启..."
        cd python_ai_service
        python3 -u ai_service.py >> ai_service.log 2>&1 &
        sleep 3
    fi
    
    if [ -n "$TRAINER_PID" ]; then
        echo "  ✅ Go Trainer: PID $TRAINER_PID"
    else
        echo "  ❌ Go Trainer: 未运行"
    fi
    
    # 检查连接
    echo ""
    echo "📡 服务状态:"
    if curl -s http://localhost:50051/health > /dev/null 2>&1; then
        HEALTH=$(curl -s http://localhost:50051/health)
        echo "  ✅ Python API: 正常"
        echo "     Buffer: $(echo $HEALTH | python3 -c "import sys,json; print(json.load(sys.stdin)['buffer_size'])" 2>/dev/null || echo "?")"
        echo "     Epsilon: $(echo $HEALTH | python3 -c "import sys,json; print(json.load(sys.stdin)['epsilon'])" 2>/dev/null || echo "?")"
        echo "     Train Count: $(echo $HEALTH | python3 -c "import sys,json; print(json.load(sys.stdin)['train_count'])" 2>/dev/null || echo "?")"
    else
        echo "  ❌ Python API: 无响应"
    fi
    
    # 统计信息
    echo ""
    echo "📈 训练统计:"
    GAMES=$(grep -c "Game over" trainer/logs/trainer-20251114.log 2>/dev/null || echo 0)
    EPISODES=$(grep -c "📦 Episode" python_ai_service/ai_service.log 2>/dev/null || echo 0)
    TRAINS=$(grep -c "🔥 Train" python_ai_service/ai_service.log 2>/dev/null || echo 0)
    CONN_REFUSED=$(grep -c "connection refused" trainer/logs/trainer-20251114.log 2>/dev/null || echo 0)
    
    echo "  游戏数: $GAMES"
    echo "  Episodes: $EPISODES"
    echo "  训练次数: $TRAINS"
    if [ "$CONN_REFUSED" -gt 0 ]; then
        echo "  ⚠️  连接失败: $CONN_REFUSED"
    fi
    
    # 最新日志
    echo ""
    echo "📝 最新训练日志:"
    tail -5 python_ai_service/ai_service.log | grep -E "(Episode|Train|Updated)" | tail -3
    
    echo ""
    echo "按 Ctrl+C 退出监控"
    echo ""
    
    sleep 10
done

