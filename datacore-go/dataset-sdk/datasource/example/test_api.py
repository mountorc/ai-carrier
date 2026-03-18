#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://localhost:8085"

def test_health():
    print("Testing health endpoint...")
    response = requests.get(f"{BASE_URL}/health")
    print(f"Status: {response.status_code}")
    print(f"Response: {response.json()}")
    print()

def test_add_datasource():
    print("Testing add datasource endpoint...")
    
    datasource_config = {
        "uuid": "test-datasource-001",
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
    
    response = requests.post(f"{BASE_URL}/datasource/add", json=payload)
    print(f"Status: {response.status_code}")
    print(f"Response: {response.json()}")
    print()

def test_query_with_uuid():
    print("Testing query with uuid_datasource endpoint...")
    
    # 先测试简单的查询
    payload = {
        "uuid_datasource": "test-datasource-001",
        "sql": "SELECT version()"
    }
    
    response = requests.post(f"{BASE_URL}/datasource/query", json=payload)
    print(f"Status: {response.status_code}")
    print(f"Response: {json.dumps(response.json(), indent=2, ensure_ascii=False)}")
    print()

if __name__ == "__main__":
    test_health()
    test_add_datasource()
    test_query_with_uuid()
