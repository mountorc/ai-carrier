#!/usr/bin/env python3
import requests
import json

DATASET_URL = "http://localhost:8084"
DATASOURCE_URL = "http://localhost:8085"

print("=" * 60)
print("测试新的 getFromDataSource API")
print("=" * 60)
print()

# 1. 首先，我们需要确保 datasource API 中有我们的数据源
print("步骤 1: 检查 datasource API 中的数据源...")

# 先测试一下直接用 datasource API 查询
print("测试 datasource API 查询 workflow 表...")
response = requests.get(f"{DATASOURCE_URL}/datasource/query/post4bc7-9a41-4332-93a1-a60c4d8a7e19?sql=SELECT%20*%20FROM%20workflow%20WHERE%20uuid%20%3D%20'workflow1'")
print(f"状态码: {response.status_code}")
if response.status_code == 200:
    data = response.json()
    print(f"响应: {json.dumps(data, indent=2, ensure_ascii=False)}")
print()

print("=" * 60)
print("API 接口说明:")
print("=" * 60)
print("新的 API 接口地址: GET /getFromDataSource")
print("参数:")
print("  - uuid_datasource: 数据源的 UUID (例如: post4bc7-9a41-4332-93a1-a60c4d8a7e19)")
print("  - uuid: 要查询的记录 UUID，格式为 uuid_表名_实际UUID (例如: uuid_workflow_workflow1)")
print()
print("使用示例:")
print("  curl -X GET 'http://localhost:8084/getFromDataSource?uuid_datasource=post4bc7-9a41-4332-93a1-a60c4d8a7e19&uuid=uuid_workflow_workflow1'")
print()
print("注意: 由于 dataset-go 和 datasource API 是两个独立的进程，")
print("需要确保在使用 /getFromDataSource 之前，已经在 dataset-go 的")
print("DataSourceManager 中添加了相应的数据源。")
print("=" * 60)
