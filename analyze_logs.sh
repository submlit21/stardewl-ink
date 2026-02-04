#!/bin/bash
echo "📊 日志分析报告"
echo "================"

# 分析每个文件的日志模式
for file in core/p2p_connector.go core/connection.go core/core.go signaling/main.go cmd/stardewl/main.go; do
    echo ""
    echo "📄 $file:"
    
    # 统计日志类型
    total=$(grep -c "log\." "$file")
    chinese=$(grep "log\." "$file" | grep -c "[一-龥]")
    english=$(grep "log\." "$file" | grep -v "[一-龥]" | grep -c "log\.")
    emoji=$(grep "log\." "$file" | grep -o -E "[🚀📞🎯✅❌⚠️🔒📡🌐💓🛑🌀]" | wc -l)
    
    echo "   总计: $total, 中文: $chinese, 英文: $english, 表情: $emoji"
    
    # 显示一些示例日志
    echo "   示例日志:"
    grep "log\." "$file" | head -3 | sed 's/^/      /'
done

echo ""
echo "🎯 建议:"
echo "   1. 统一使用英文日志（更国际化）"
echo "   2. 减少表情符号使用（影响日志解析）"
echo "   3. 移除过于详细的调试日志"
