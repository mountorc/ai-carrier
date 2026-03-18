#!/usr/bin/env python3
content = '''#!/usr/bin/env python3
import requests
import json

DATASET_URL = "http://localhost:8084"

print("=" * 60)
print("测试新的 getAutoSet API")
print("=" * 60)
print()

# 1. 首先添加数据源
print("步骤 1: 添加数据源...")
datasource_config = {
    "uuid": "post4bc7-9a41-4332-93a1-a60c4d8a7e19",
    "type": "postgres",
    "config": {
        "host": "121.43.142.153",
        "port": 5432,
        "database": "carrier",
        "username": "carrier",
        "password": "GNerfiSP4dpZjwcJ"
    }
}

payload = {
    "config_json": json.dumps(datasource_config)
}

response = requests.post(f"{DATASET_URL}/datasource/add", json=payload)
print(f"添加数据源状态码: {response.status_code}")
print(f"响应: {response.json()}")
print()

# 2. 测试新的 getAutoSet API
print("步骤 2: 测试 getAutoSet API (uuid_datasource + uuid_workflow)...")
response = requests.get(f"{DATASET_URL}/getAutoSet", params={
    "uuid_datasource": "post4bc7-9a41-4332-93a1-a60c4d8a7e19",
    "uuid_workflow": "workflow1"
})
print(f"状态码: {response.status_code}")
if response.status_code == 200:
    print(f"响应: {json.dumps(response.json(), indent=2, ensure_ascii=False)}")
else:
    print(f"响应文本: {response.text}")
print()

print("=" * 60)
print("API 测试完成！")
print("=" * 60)
print()
print("使用示例:")
print("  curl -X GET 'http://localhost:8084/getAutoSet?uuid_datasource=post4bc7-9a41-4332-93a1-a60c4d8a7e19&uuid_workflow=workflow1'")
print("  curl -X GET 'http://localhost:8084/getAutoSet?uuid_datasource=post4bc7-9a41-4332-93a1-a60c4d8a7e19&uuid_dataset=xxx'")
print("=" * 60)
'''

with open('/Users/a1-6/Documents/code/JavaProject/autoDataSource/datacore-go/dataset-go/test_new_getautoset.py', 'w') as f:
    f.write(content)

print("Test file written successfully")
