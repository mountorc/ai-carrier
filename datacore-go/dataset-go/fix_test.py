#!/usr/bin/env python3
content = '''#!/usr/bin/env python3
import requests
import json

DATASET_URL = "http://localhost:8084"
DATASOURCE_URL = "http://localhost:8085"

def test_health_dataset():
    print("Testing dataset-go health endpoint...")
    response = requests.get(f"{DATASET_URL}/health")
    print(f"Status: {response.status_code}")
    print(f"Response: {response.json()}")
    print()

def test_add_datasource():
    print("Adding test datasource to datasource API...")
    
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
    
    response = requests.post(f"{DATASOURCE_URL}/datasource/add", json=payload)
    print(f"Status: {response.status_code}")
    print(f"Response: {response.json()}")
    print()

def test_query_datasource():
    print("Testing datasource query endpoint...")
    response = requests.get(f"{DATASOURCE_URL}/datasource/query/post4bc7-9a41-4332-93a1-a60c4d8a7e19?sql=SELECT%20version%28%29")
    print(f"Status: {response.status_code}")
    print(f"Response: {json.dumps(response.json(), indent=2, ensure_ascii=False)}")
    print()

def test_get_from_datasource():
    print("Testing new getFromDataSource endpoint (dataset-go)...")
    
    # 先查询一下 dataset 表看看有什么数据
    print("First, query dataset table...")
    response = requests.get(f"{DATASOURCE_URL}/datasource/query/post4bc7-9a41-4332-93a1-a60c4d8a7e19?sql=SELECT%20uuid%20FROM%20dataset%20LIMIT%201")
    print(f"Dataset query status: {response.status_code}")
    
    if response.status_code == 200:
        data = response.json()
        print(f"Dataset query response: {json.dumps(data, indent=2, ensure_ascii=False)}")
        
        if data.get("data") and len(data["data"]) > 0:
            first_uuid = data["data"][0].get("uuid")
            print(f"\\nFound dataset uuid: {first_uuid}")
            
            # 用新的 API 来查询
            print(f"\\nTesting getFromDataSource with uuid_dataset...")
            response2 = requests.get(f"{DATASET_URL}/getFromDataSource", params={
                "uuid_datasource": "post4bc7-9a41-4332-93a1-a60c4d8a7e19",
                "uuid": f"uuid_dataset_{first_uuid}"
            })
            print(f"Status: {response2.status_code}")
            print(f"Response: {json.dumps(response2.json(), indent=2, ensure_ascii=False)}")

if __name__ == "__main__":
    test_health_dataset()
    test_add_datasource()
    test_query_datasource()
    test_get_from_datasource()
'''

with open('/Users/a1-6/Documents/code/JavaProject/autoDataSource/datacore-go/dataset-go/test_new_api.py', 'w') as f:
    f.write(content)

print("Test file written successfully")
