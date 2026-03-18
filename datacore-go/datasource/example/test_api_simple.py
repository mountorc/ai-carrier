#!/usr/bin/env python3
import requests
import json
import time

BASE_URL = "http://localhost:8085"

print("=" * 60)
print("Datasource API Test")
print("=" * 60)

# Test 1: Health Check
print("\n1. Testing Health Check...")
try:
    response = requests.get(f"{BASE_URL}/health")
    print(f"   Status: {response.status_code}")
    print(f"   Response: {json.dumps(response.json(), indent=2)}")
except Exception as e:
    print(f"   Error: {e}")

time.sleep(1)

# Test 2: Add Datasource
print("\n2. Testing Add Datasource...")
config = {
    "uuid": "uuid_datasource_post4bc7-9a41-4332-93a1-a60c4d8a7e19",
    "type": "postgres",
    "config": {
        "host": "http://121.43.142.153",
        "port": 5432,
        "charset": "utf-8",
        "database": "carrier",
        "password": "GNerfiSP4dpZjwcJ",
        "username": "carrier",
        "driver_class": "org.postgresql.Driver"
    }
}

payload = {
    "config_json": json.dumps(config)
}

try:
    response = requests.post(
        f"{BASE_URL}/datasource/add",
        json=payload,
        headers={"Content-Type": "application/json"}
    )
    print(f"   Status: {response.status_code}")
    print(f"   Response: {json.dumps(response.json(), indent=2)}")
except Exception as e:
    print(f"   Error: {e}")

time.sleep(1)

# Test 3: Query with UUID
print("\n3. Testing Query with UUID...")
query_payload = {
    "uuid_datasource": "uuid_datasource_post4bc7-9a41-4332-93a1-a60c4d8a7e19",
    "sql": "SELECT version()"
}

try:
    response = requests.post(
        f"{BASE_URL}/datasource/query",
        json=query_payload,
        headers={"Content-Type": "application/json"}
    )
    print(f"   Status: {response.status_code}")
    print(f"   Response: {json.dumps(response.json(), indent=2)}")
except Exception as e:
    print(f"   Error: {e}")

time.sleep(1)

# Test 4: Query Current Database
print("\n4. Testing Query Current Database...")
query_payload2 = {
    "uuid_datasource": "uuid_datasource_post4bc7-9a41-4332-93a1-a60c4d8a7e19",
    "sql": "SELECT current_database()"
}

try:
    response = requests.post(
        f"{BASE_URL}/datasource/query",
        json=query_payload2,
        headers={"Content-Type": "application/json"}
    )
    print(f"   Status: {response.status_code}")
    print(f"   Response: {json.dumps(response.json(), indent=2)}")
except Exception as e:
    print(f"   Error: {e}")

print("\n" + "=" * 60)
print("Test Complete!")
print("=" * 60)
