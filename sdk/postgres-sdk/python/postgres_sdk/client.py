import psycopg2
from psycopg2.extras import RealDictCursor
import json
import requests
from typing import List, Dict, Any, Optional, Tuple


class PostgresClient:
    def __init__(self, embedding_api_url: Optional[str] = None, 
                 use_aliyun: bool = True, 
                 aliyun_api_key: Optional[str] = None,
                 aliyun_model: str = "text-embedding-v4"):
        """
        初始化PostgreSQL客户端
        
        Args:
            embedding_api_url: 可选的embedding API地址，如果设置则使用远程API而不是本地模型
            use_aliyun: 是否使用阿里云百炼API
            aliyun_api_key: 阿里云API Key
            aliyun_model: 阿里云模型名称，默认为text-embedding-v4
        """
        self.connection = None
        self.connection_params = None
        self.embedding_api_url = embedding_api_url
        self.use_aliyun = use_aliyun
        self.aliyun_api_key = aliyun_api_key
        self.aliyun_model = aliyun_model

    def connect_from_url(self, db_url: str):
        self.connection = psycopg2.connect(db_url)
        self.connection.autocommit = True
        print(f"Connected to PostgreSQL at {db_url}")
        return self

    def connect(self, host: str, port: int, database: str, user: str, password: str):
        self.connection_params = {
            'host': host,
            'port': port,
            'database': database,
            'user': user,
            'password': password
        }
        self.connection = psycopg2.connect(**self.connection_params)
        self.connection.autocommit = True
        print(f"Connected to PostgreSQL at {host}:{port}/{database}")
        return self

    def disconnect(self):
        if self.connection:
            self.connection.close()
            print("Disconnected from PostgreSQL")

    def create_collection(self, collection_name: str, dimension: int, index_type: str = None):
        with self.connection.cursor(cursor_factory=RealDictCursor) as cursor:
            create_table_sql = f"""
                CREATE TABLE IF NOT EXISTS {collection_name} (
                    id SERIAL PRIMARY KEY,
                    embedding vector({dimension}),
                    metadata jsonb
                )
            """
            cursor.execute(create_table_sql)

            if index_type:
                if index_type.upper() == 'IVFFLAT':
                    create_index_sql = f"""
                        CREATE INDEX IF NOT EXISTS {collection_name}_embedding_idx
                        ON {collection_name} USING ivfflat (embedding vector_l2_ops)
                        WITH (lists = 128)
                    """
                elif index_type.upper() == 'HNSW':
                    create_index_sql = f"""
                        CREATE INDEX IF NOT EXISTS {collection_name}_embedding_idx
                        ON {collection_name} USING hnsw (embedding vector_l2_ops)
                        WITH (m = 16, ef_construction = 64)
                    """
                else:
                    raise ValueError(f"Unsupported index type: {index_type}")
                cursor.execute(create_index_sql)

        print(f"Collection {collection_name} created with dimension {dimension}")

    def drop_collection(self, collection_name: str):
        with self.connection.cursor() as cursor:
            cursor.execute(f"DROP TABLE IF EXISTS {collection_name}")
        print(f"Collection {collection_name} dropped")

    def list_collections(self) -> List[str]:
        with self.connection.cursor(cursor_factory=RealDictCursor) as cursor:
            cursor.execute("""
                SELECT table_name FROM information_schema.tables
                WHERE table_schema = 'public'
            """)
            return [row['table_name'] for row in cursor.fetchall()]

    def insert_vectors(self, collection_name: str, vectors: List[List[float]],
                       metadata: Optional[List[Dict[str, Any]]] = None):
        if not metadata:
            metadata = [{} for _ in range(len(vectors))]

        with self.connection.cursor() as cursor:
            for i, vector in enumerate(vectors):
                vector_str = '[' + ', '.join([str(x) for x in vector]) + ']'
                metadata_json = json.dumps(metadata[i])
                cursor.execute(
                    f"INSERT INTO {collection_name} (embedding, metadata) VALUES (%s::vector, %s::jsonb)",
                    (vector_str, metadata_json)
                )
        print(f"Inserted {len(vectors)} vectors into {collection_name}")

    def search_vectors(self, collection_name: str, query_vector: List[float],
                       top_k: int = 10, filter_condition: Optional[str] = None) -> List[Dict[str, Any]]:
        vector_str = '[' + ', '.join([str(x) for x in query_vector]) + ']'

        sql = f"""
            SELECT id, embedding <-> %s::vector as distance, metadata
            FROM {collection_name}
        """
        if filter_condition:
            sql += f" WHERE {filter_condition}"
        sql += " ORDER BY distance LIMIT %s"

        with self.connection.cursor(cursor_factory=RealDictCursor) as cursor:
            cursor.execute(sql, (vector_str, top_k))
            results = []
            for row in cursor.fetchall():
                results.append({
                    'id': row['id'],
                    'score': float(row['distance']),
                    'metadata': row['metadata']
                })
            return results

    def execute_sql(self, sql: str, params: Tuple = None) -> List[Dict[str, Any]]:
        with self.connection.cursor(cursor_factory=RealDictCursor) as cursor:
            cursor.execute(sql, params or ())
            if cursor.description:
                return [dict(row) for row in cursor.fetchall()]
            return []

    def get_table_fields(self, table_schema: str, table_name: str) -> List[Dict[str, Any]]:
        sql = """
            SELECT column_name, data_type, character_maximum_length
            FROM information_schema.columns
            WHERE table_schema = %s AND table_name = %s
            ORDER BY ordinal_position
        """
        return self.execute_sql(sql, (table_schema, table_name))

    def get_table_list(self, table_schema: str = 'public') -> List[str]:
        sql = """
            SELECT table_name
            FROM information_schema.tables
            WHERE table_schema = %s
            ORDER BY table_name
        """
        results = self.execute_sql(sql, (table_schema,))
        return [row['table_name'] for row in results]

    def begin_transaction(self):
        self.connection.autocommit = False

    def commit(self):
        self.connection.commit()
        self.connection.autocommit = True

    def rollback(self):
        self.connection.rollback()
        self.connection.autocommit = True
    
    def _l2_distance(self, v1, v2):
        """计算L2距离"""
        import math
        return math.sqrt(sum((x - y) ** 2 for x, y in zip(v1, v2)))
    
    def _cosine_similarity(self, v1, v2):
        """计算余弦相似度"""
        import math
        dot_product = sum(x * y for x, y in zip(v1, v2))
        norm1 = math.sqrt(sum(x ** 2 for x in v1))
        norm2 = math.sqrt(sum(y ** 2 for y in v2))
        if norm1 == 0 or norm2 == 0:
            return 0.0
        return dot_product / (norm1 * norm2)
    
    def _inner_product(self, v1, v2):
        """计算内积"""
        return sum(x * y for x, y in zip(v1, v2))
    
    def search_jsonb_vectors(self, table_name: str, query_vector: List[float], 
                            vector_column: str = 'vec_embedding', 
                            top_k: int = 10, 
                            distance_type: str = 'l2',
                            select_columns: Optional[List[str]] = None) -> List[Dict[str, Any]]:
        """
        搜索JSONB格式存储的向量
        
        Args:
            table_name: 表名
            query_vector: 查询向量
            vector_column: 向量列名，默认为'vec_embedding'
            top_k: 返回结果数量
            distance_type: 距离类型，可选'l2'、'cosine'、'inner_product'
            select_columns: 选择的列名列表，默认为None（选择所有列）
        
        Returns:
            搜索结果列表，包含distance字段
        """
        if select_columns:
            columns_sql = ', '.join(select_columns)
        else:
            columns_sql = '*'
        
        sql = f"SELECT {columns_sql}, {vector_column} FROM {table_name}"
        
        with self.connection.cursor(cursor_factory=RealDictCursor) as cursor:
            cursor.execute(sql)
            all_rows = cursor.fetchall()
        
        results = []
        for row in all_rows:
            vec_data = row.get(vector_column)
            if vec_data:
                if isinstance(vec_data, str):
                    vec = json.loads(vec_data)
                else:
                    vec = vec_data
                
                if distance_type == 'l2':
                    distance = self._l2_distance(query_vector, vec)
                elif distance_type == 'cosine':
                    similarity = self._cosine_similarity(query_vector, vec)
                    distance = 1 - similarity
                elif distance_type == 'inner_product':
                    distance = -self._inner_product(query_vector, vec)
                else:
                    raise ValueError(f"Unsupported distance type: {distance_type}")
                
                result_row = dict(row)
                result_row['distance'] = distance
                del result_row[vector_column]
                results.append(result_row)
        
        results.sort(key=lambda x: x['distance'])
        return results[:top_k]
    
    def get_embedding(self, text: str, model_name: str = 'all-MiniLm-L12-v2') -> List[float]:
        """
        获取文本的embedding向量
        
        Args:
            text: 输入文本
            model_name: sentence-transformers模型名称
        
        Returns:
            embedding向量
        """
        if self.use_aliyun:
            from .embedding import AliyunEmbedding
            embedding = AliyunEmbedding(api_key=self.aliyun_api_key, model=self.aliyun_model)
            return embedding.embed_text(text)
        elif self.embedding_api_url:
            return self._get_embedding_from_api(text)
        else:
            from .embedding import PostgresEmbedding
            embedding = PostgresEmbedding(model_name)
            return embedding.embed_text(text)
    
    def _get_embedding_from_api(self, text: str) -> List[float]:
        """
        从远程API获取embedding
        
        Args:
            text: 输入文本
        
        Returns:
            embedding向量
        """
        if not self.embedding_api_url:
            raise ValueError("embedding_api_url not set")
        
        url = f"{self.embedding_api_url.rstrip('/')}/embed"
        response = requests.get(url, params={'text': text})
        
        if response.status_code != 200:
            raise Exception(f"Embedding API error: {response.status_code}")
        
        result = response.json()
        if not result.get('success'):
            raise Exception(f"Embedding API failed: {result.get('error')}")
        
        return result.get('embedding')
    
    def search_by_text(self, table_name: str, query_text: str,
                      vector_column: str = 'vec_embedding',
                      top_k: int = 10,
                      distance_type: str = 'l2',
                      select_columns: Optional[List[str]] = None,
                      model_name: str = 'all-MiniLm-L12-v2') -> List[Dict[str, Any]]:
        """
        通过文本搜索（自动生成embedding）
        
        Args:
            table_name: 表名
            query_text: 查询文本
            vector_column: 向量列名，默认为'vec_embedding'
            top_k: 返回结果数量
            distance_type: 距离类型，可选'l2'、'cosine'、'inner_product'
            select_columns: 选择的列名列表，默认为None（选择所有列）
            model_name: sentence-transformers模型名称
        
        Returns:
            搜索结果列表
        """
        query_vector = self.get_embedding(query_text, model_name)
        return self.search_jsonb_vectors(
            table_name, query_vector, vector_column, 
            top_k, distance_type, select_columns
        )
