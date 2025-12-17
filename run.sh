#!/bin/bash

# FF14 捡垃圾助手管理脚本
APP_NAME="llmaget"
APP_DIR=$(cd "$(dirname "$0")" && pwd)
APP_PATH="$APP_DIR/$APP_NAME"
PID_FILE="$APP_DIR/crawler.pid"
LOG_FILE="$APP_DIR/crawler.log"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查应用是否存在
check_app() {
    if [ ! -f "$APP_PATH" ]; then
        echo -e "${RED}错误: $APP_PATH 不存在${NC}"
        exit 1
    fi
    if [ ! -x "$APP_PATH" ]; then
        chmod +x "$APP_PATH"
    fi
}

# 获取PID
get_pid() {
    if [ -f "$PID_FILE" ]; then
        cat "$PID_FILE"
    else
        echo ""
    fi
}

# 检查进程是否运行
is_running() {
    local pid=$(get_pid)
    if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# 启动服务
start() {
    check_app
    
    if is_running; then
        echo -e "${YELLOW}服务已在运行中 (PID: $(get_pid))${NC}"
        return 0
    fi
    
    echo -e "${GREEN}启动服务...${NC}"
    cd "$APP_DIR"
    nohup "$APP_PATH" >> "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    
    sleep 1
    if is_running; then
        echo -e "${GREEN}✓ 服务启动成功 (PID: $(get_pid))${NC}"
        echo -e "  日志文件: $LOG_FILE"
        echo -e "  接口地址: http://localhost:8080/llmaget/ff_info"
    else
        echo -e "${RED}✗ 服务启动失败，请查看日志: $LOG_FILE${NC}"
        rm -f "$PID_FILE"
        exit 1
    fi
}

# 停止服务
stop() {
    if ! is_running; then
        echo -e "${YELLOW}服务未运行${NC}"
        rm -f "$PID_FILE"
        return 0
    fi
    
    local pid=$(get_pid)
    echo -e "${GREEN}停止服务 (PID: $pid)...${NC}"
    kill "$pid"
    
    # 等待进程退出
    for i in {1..10}; do
        if ! is_running; then
            echo -e "${GREEN}✓ 服务已停止${NC}"
            rm -f "$PID_FILE"
            return 0
        fi
        sleep 1
    done
    
    # 强制终止
    echo -e "${YELLOW}强制终止进程...${NC}"
    kill -9 "$pid" 2>/dev/null
    rm -f "$PID_FILE"
    echo -e "${GREEN}✓ 服务已停止${NC}"
}

# 重启服务
restart() {
    stop
    sleep 1
    start
}

# 查看状态
status() {
    if is_running; then
        echo -e "${GREEN}✓ 服务运行中 (PID: $(get_pid))${NC}"
        echo -e "  日志文件: $LOG_FILE"
        echo -e "  接口地址: http://localhost:8080/llmaget/ff_info"
    else
        echo -e "${RED}✗ 服务未运行${NC}"
    fi
}

# 查看日志
logs() {
    if [ -f "$LOG_FILE" ]; then
        tail -f "$LOG_FILE"
    else
        echo -e "${YELLOW}日志文件不存在: $LOG_FILE${NC}"
    fi
}

# 查看最近日志
tail_logs() {
    local lines=${1:-50}
    if [ -f "$LOG_FILE" ]; then
        tail -n "$lines" "$LOG_FILE"
    else
        echo -e "${YELLOW}日志文件不存在: $LOG_FILE${NC}"
    fi
}

# 帮助信息
usage() {
    echo "FF14 捡垃圾助手管理脚本"
    echo ""
    echo "用法: $0 {start|stop|restart|status|logs|tail [行数]}"
    echo ""
    echo "命令:"
    echo "  start    启动服务"
    echo "  stop     停止服务"
    echo "  restart  重启服务"
    echo "  status   查看服务状态"
    echo "  logs     实时查看日志 (Ctrl+C 退出)"
    echo "  tail     查看最近日志 (默认50行)"
    echo ""
}

# 主逻辑
case "$1" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    logs)
        logs
        ;;
    tail)
        tail_logs "$2"
        ;;
    *)
        usage
        exit 1
        ;;
esac

exit 0


