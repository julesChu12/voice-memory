#!/bin/bash
API="http://localhost:8080"

echo "=== 测试会话ID关联功能 ==="
echo ""

# 1. 创建新会话并获取session_id
echo "1. 发送消息创建会话..."
RESP=$(curl -s -X POST "${API}/api/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"测试消息1","session_id":""}')

SESSION_ID=$(echo $RESP | jq -r '.session_id // .sessionID // empty')
echo "Session ID: $SESSION_ID"
echo ""

if [ -z "$SESSION_ID" ]; then
  echo "❌ 获取session_id失败，响应："
  echo $RESP
  # 使用手动生成的session_id
  SESSION_ID="session_test_$(date +%s)"
  echo "使用手动生成的 session_id: $SESSION_ID"
fi

# 2. 保存知识（关联session_id）
echo "2. 保存知识并关联会话..."
KNOW_RESP=$(curl -s -X POST "${API}/api/knowledge/record" \
  -F "text=这是测试知识，应该关联到会话 ${SESSION_ID}" \
  -F "session_id=${SESSION_ID}" \
  -F "auto_organize=true")

echo $KNOW_RESP | jq '.'
echo ""

# 3. 查看知识列表，检查session_id
echo "3. 验证知识列表中的session_id..."
curl -s "${API}/api/knowledge/list" | jq '.knowledges[-1] | {id, content, session_id}'
echo ""

# 4. 查看所有会话
echo "4. 查看所有会话..."
curl -s "${API}/api/sessions" | jq '.sessions | length'
echo ""

echo "=== 测试完成 ==="
