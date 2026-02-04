#!/bin/bash

# Stardewl-Ink 测试运行器
# 用法: ./run_tests.sh [类型] [测试名]
#   类型: all, unit, integration, e2e
#   测试名: 可选，指定具体测试文件

set -e

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$BASE_DIR/.." && pwd)"
DIST_DIR="$PROJECT_ROOT/dist"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查构建
check_build() {
    if [ ! -f "$DIST_DIR/stardewl" ] || [ ! -f "$DIST_DIR/stardewl-signaling" ]; then
        log_warning "构建文件不存在，正在构建..."
        cd "$PROJECT_ROOT"
        make build
        if [ $? -ne 0 ]; then
            log_error "构建失败"
            exit 1
        fi
        log_success "构建完成"
    fi
}

# 清理函数
cleanup() {
    log_info "清理进程..."
    pkill -f stardewl 2>/dev/null || true
    pkill -f stardewl-signaling 2>/dev/null || true
    sleep 1
}

# 运行单个测试
run_single_test() {
    local test_file="$1"
    local test_name="$(basename "$test_file")"
    
    log_info "运行测试: $test_name"
    
    # 确保可执行
    chmod +x "$test_file"
    
    # 运行测试
    cd "$PROJECT_ROOT"
    timeout 30 "$test_file"
    local result=$?
    
    if [ $result -eq 0 ]; then
        log_success "测试通过: $test_name"
        return 0
    elif [ $result -eq 124 ]; then
        log_warning "测试超时: $test_name"
        return 1
    else
        log_error "测试失败: $test_name (退出码: $result)"
        return 1
    fi
}

# 运行单元测试
run_unit_tests() {
    log_info "运行单元测试..."
    local passed=0
    local failed=0
    
    for test_file in "$BASE_DIR/unit/"*.sh; do
        if [ -f "$test_file" ]; then
            if run_single_test "$test_file"; then
                ((passed++))
            else
                ((failed++))
            fi
        fi
    done
    
    log_info "单元测试结果: 通过 $passed, 失败 $failed"
    return $failed
}

# 运行集成测试
run_integration_tests() {
    log_info "运行集成测试..."
    local passed=0
    local failed=0
    
    # 先启动信令服务器
    log_info "启动信令服务器..."
    cd "$PROJECT_ROOT"
    ./dist/stardewl-signaling &
    local server_pid=$!
    sleep 3
    
    # 检查服务器是否运行
    if ! kill -0 $server_pid 2>/dev/null; then
        log_error "信令服务器启动失败"
        return 1
    fi
    log_success "信令服务器已启动 (PID: $server_pid)"
    
    # 运行测试
    for test_file in "$BASE_DIR/integration/"*.sh; do
        if [ -f "$test_file" ]; then
            cleanup  # 清理之前的进程
            if run_single_test "$test_file"; then
                ((passed++))
            else
                ((failed++))
            fi
        fi
    done
    
    # 停止服务器
    log_info "停止信令服务器..."
    kill $server_pid 2>/dev/null || true
    wait $server_pid 2>/dev/null || true
    
    log_info "集成测试结果: 通过 $passed, 失败 $failed"
    return $failed
}

# 运行端到端测试
run_e2e_tests() {
    log_info "运行端到端测试..."
    local passed=0
    local failed=0
    
    for test_file in "$BASE_DIR/e2e/"*.sh; do
        if [ -f "$test_file" ]; then
            cleanup  # 清理之前的进程
            if run_single_test "$test_file"; then
                ((passed++))
            else
                ((failed++))
            fi
        fi
    done
    
    log_info "端到端测试结果: 通过 $passed, 失败 $failed"
    return $failed
}

# 运行所有测试
run_all_tests() {
    log_info "运行所有测试..."
    local total_failed=0
    
    check_build
    cleanup
    
    # 运行单元测试
    if ! run_unit_tests; then
        total_failed=$((total_failed + $?))
    fi
    
    # 运行集成测试
    if ! run_integration_tests; then
        total_failed=$((total_failed + $?))
    fi
    
    # 运行端到端测试
    if ! run_e2e_tests; then
        total_failed=$((total_failed + $?))
    fi
    
    cleanup
    
    if [ $total_failed -eq 0 ]; then
        log_success "所有测试通过!"
        return 0
    else
        log_error "有 $total_failed 个测试失败"
        return 1
    fi
}

# 主函数
main() {
    local test_type="${1:-all}"
    local specific_test="${2:-}"
    
    log_info "Stardewl-Ink 测试套件"
    log_info "====================="
    
    check_build
    
    case "$test_type" in
        "all")
            run_all_tests
            ;;
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "e2e")
            run_e2e_tests
            ;;
        "single")
            if [ -z "$specific_test" ]; then
                log_error "请指定测试文件"
                exit 1
            fi
            run_single_test "$specific_test"
            ;;
        *)
            log_error "未知的测试类型: $test_type"
            echo "用法: $0 [类型] [测试名]"
            echo "类型: all, unit, integration, e2e, single"
            exit 1
            ;;
    esac
    
    local result=$?
    cleanup
    exit $result
}

# 捕获退出信号
trap cleanup EXIT INT TERM

# 运行主函数
main "$@"