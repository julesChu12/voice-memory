#!/bin/bash

# 测试会话关联功能
API_BASE="http://localhost:8080"

echo "=== 测试会话关联功能 ==="
echo ""

# 1. 创建会话
echo "1. 创建新会话..."
SESSION_RESP=$(curl -s -X POST "${API_BASE}/api/chat" \
  -H "Content-Type: application/json" \
  -d '{"message":"你好，我是测试用户","session_id":""}')

SESSION_ID=$(echo $SESSION_RESP | jq -r '.session_id // empty')
echo "   Session ID: $SESSION_ID"
echo ""

if [ -z "$SESSION_ID" ]; then
  echo "❌ 创建会话失败"
  exit 1
fi

# 2. 发送第二条消息
echo "2. 发送第二条消息..."
curl -s -X POST "${API_BASE}/api/chat" \
  -H "Content-Type: application/json" \
  -d "{\"message\":\"今天天气怎么样\",\"session_id\":\"$SESSION_ID\"}" | jq '.reply'
echo ""

# 3. 保存知识（关联会话）
echo "3. 保存知识并关联会话..."
KNOW_RESP=$(curl -s -X POST "${API_BASE}/api/knowledge/record" \
  -F "text=测试保存知识功能，应该关联到会话$SESSION_ID" \
  -F "session_id=$SESSION_ID" \
  -F "auto_organize=true")

echo $KNOW_RESP | jq '.'
KNOW_ID=$(echo $KNOW_RESP | jq -r '.knowledge.id // empty')
echo "   Knowledge ID: $KNOW_ID"
echo ""

# 4. 查看知识列表
echo "4. 查看知识列表..."
curl -s "${API_BASE}/api/knowledge/list" | jq '.knowledges[] | {id, content, session_id}'
echo ""

# 5. 查看关联的会话
echo "5. 查看知识关联的会话..."
if [ -n "$SESSION_ID" ]; then
  curl -s "${API_BASE}/api/sessions/get?session_id=$SESSION_ID" | jq '.session.messages[] | {role, content}'
fi
echo ""

# 6. 查看所有会话
echo "6. 查看所有会话..."
curl -s "${API_BASE}/api/sessions" | jq '.sessions[] | {id, message_count: (.messages | length)}'
echo ""

echo "=== 测试完成 ==="
