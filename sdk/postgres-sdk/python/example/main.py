#!/usr/bin/env python3
import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from postgres_sdk import PostgresClient


def main():
    client = PostgresClient()
    
    try:
        db_url = "postgresql://carrier:GNerfiSP4dpZjwcJ@121.43.142.153:5432/carrier"
        print(f"Connecting to database...")
        client.connect_from_url(db_url)
        
        print("\n=== List collections ===")
        collections = client.list_collections()
        print(f"Collections: {collections}")
        
        collection_name = "test_vectors_sdk"
        dimension = 128
        
        print(f"\n=== Create collection {collection_name} with dimension {dimension} ===")
        client.create_collection(collection_name, dimension, index_type="IVFFLAT")
        
        print("\n=== Insert vectors ===")
        vectors = [
            [0.1 * i for i in range(dimension)],
            [0.2 * i for i in range(dimension)],
            [0.3 * i for i in range(dimension)],
        ]
        metadata = [
            {"text": "Hello world", "category": "test"},
            {"text": "Python SDK", "category": "sdk"},
            {"text": "PostgreSQL vector", "category": "database"},
        ]
        client.insert_vectors(collection_name, vectors, metadata)
        
        print("\n=== Search vectors ===")
        query_vector = [0.15 * i for i in range(dimension)]
        results = client.search_vectors(collection_name, query_vector, top_k=3)
        for result in results:
            print(f"ID: {result['id']}, Score: {result['score']:.4f}, Metadata: {result['metadata']}")
        
        print("\n=== Get table list ===")
        tables = client.get_table_list()
        print(f"Tables: {tables}")
        
        if tables:
            print(f"\n=== Get fields for table {tables[0]} ===")
            fields = client.get_table_fields('public', tables[0])
            print(f"Fields: {fields}")
        
        print("\n=== Drop collection ===")
        client.drop_collection(collection_name)
        
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()
    finally:
        client.disconnect()


if __name__ == "__main__":
    main()
