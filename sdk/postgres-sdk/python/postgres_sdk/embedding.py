"""
PostgreSQL Embedding Module
支持sentence-transformers和阿里云百炼API进行文本embedding
"""

import os
import sys
import json
from typing import List, Optional, Dict, Any

try:
    from sentence_transformers import SentenceTransformer
    import numpy as np
    SENTENCE_TRANSFORMERS_AVAILABLE = True
except ImportError:
    SENTENCE_TRANSFORMERS_AVAILABLE = False
    print("警告: sentence-transformers 未安装，请先运行: pip install sentence-transformers")

try:
    import requests
    REQUESTS_AVAILABLE = True
except ImportError:
    REQUESTS_AVAILABLE = False
    print("警告: requests 未安装，请先运行: pip install requests")


def get_local_model_path():
    """获取本地模型路径"""
    current_dir = os.path.dirname(os.path.abspath(__file__))
    model_dir = os.path.join(current_dir, '../../models', 'all-MiniLm-L12-v2')
    return os.path.abspath(model_dir)


class AliyunEmbedding:
    """阿里云百炼 Embedding 类"""
    
    DEFAULT_API_URL = "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings"
    DEFAULT_MODEL = "text-embedding-v4"
    DEFAULT_API_KEY = "sk-403ca84daa9740df82ce0a1737ceccdf"
    
    def __init__(self, api_key: Optional[str] = None, model: str = DEFAULT_MODEL, api_url: str = DEFAULT_API_URL):
        """
        初始化阿里云百炼 Embedding
        
        Args:
            api_key: 阿里云API Key，默认为sk-403ca84daa9740df82ce0a1737ceccdf
            model: 模型名称，默认为text-embedding-v4
            api_url: API地址，默认为阿里云百炼地址
        """
        if not REQUESTS_AVAILABLE:
            raise ImportError("requests 未安装，请先安装: pip install requests")
        
        self.api_key = api_key or self.DEFAULT_API_KEY
        self.model = model
        self.api_url = api_url
        self._dimension = None
        print(f"初始化阿里云百炼Embedding: 模型={model}")
    
    def embed_text(self, text: str) -> List[float]:
        """
        将文本转换为embedding向量
        
        Args:
            text: 输入文本
            
        Returns:
            embedding向量
        """
        headers = {
            "Authorization": f"Bearer {self.api_key}",
            "Content-Type": "application/json"
        }
        
        data = {
            "model": self.model,
            "input": text
        }
        
        response = requests.post(self.api_url, headers=headers, json=data)
        
        if response.status_code != 200:
            raise Exception(f"阿里云API调用失败: {response.status_code}, {response.text}")
        
        result = response.json()
        
        if "data" not in result or not result["data"]:
            raise Exception(f"阿里云API返回格式错误: {result}")
        
        embedding = result["data"][0]["embedding"]
        
        if self._dimension is None:
            self._dimension = len(embedding)
            print(f"模型维度: {self._dimension}")
        
        return embedding
    
    def embed_texts(self, texts: List[str]) -> List[List[float]]:
        """
        批量将文本转换为embedding向量
        
        Args:
            texts: 输入文本列表
            
        Returns:
            embedding向量列表
        """
        results = []
        for text in texts:
            results.append(self.embed_text(text))
        return results
    
    def get_embedding_dimension(self) -> int:
        """获取embedding向量维度"""
        if self._dimension is None:
            self.embed_text("test")
        return self._dimension


class PostgresEmbedding:
    """PostgreSQL Embedding 类，支持sentence-transformers和阿里云百炼"""
    
    def __init__(self, model_name: str = "all-MiniLm-L12-v2", use_local: bool = True, use_aliyun: bool = False, 
                 aliyun_api_key: Optional[str] = None, aliyun_model: str = AliyunEmbedding.DEFAULT_MODEL):
        """
        初始化PostgreSQL Embedding
        
        Args:
            model_name: sentence-transformers模型名称，默认为all-MiniLm-L12-v2
            use_local: 是否优先使用本地模型
            use_aliyun: 是否使用阿里云百炼API
            aliyun_api_key: 阿里云API Key
            aliyun_model: 阿里云模型名称
        """
        self.use_aliyun = use_aliyun
        
        if use_aliyun:
            self._aliyun_embedding = AliyunEmbedding(api_key=aliyun_api_key, model=aliyun_model)
        else:
            if not SENTENCE_TRANSFORMERS_AVAILABLE:
                raise ImportError("sentence-transformers 未安装，请先安装: pip install sentence-transformers")
            
            self.model_name = model_name
            self.use_local = use_local
            self.model = None
            
            self._init_model()
    
    def _init_model(self):
        """初始化embedding模型"""
        local_model_path = get_local_model_path()
        
        if self.use_local and os.path.exists(local_model_path):
            print(f"正在从本地加载embedding模型: {local_model_path}")
            self.model = SentenceTransformer(local_model_path)
            print(f"本地模型加载完成！向量维度: {self.model.get_sentence_embedding_dimension()}")
        else:
            print(f"正在加载embedding模型: {self.model_name}")
            print("首次使用会自动下载模型，请稍候...")
            self.model = SentenceTransformer(self.model_name)
            print(f"模型加载完成！向量维度: {self.model.get_sentence_embedding_dimension()}")
            
            if self.use_local:
                print(f"\n正在保存模型到本地: {local_model_path}")
                os.makedirs(os.path.dirname(local_model_path), exist_ok=True)
                self.model.save(local_model_path)
                print(f"模型已保存到本地！")
    
    def embed_text(self, text: str) -> List[float]:
        """
        将文本转换为embedding向量
        
        Args:
            text: 输入文本
            
        Returns:
            embedding向量
        """
        if self.use_aliyun:
            return self._aliyun_embedding.embed_text(text)
        return self.model.encode(text).tolist()
    
    def embed_texts(self, texts: List[str]) -> List[List[float]]:
        """
        批量将文本转换为embedding向量
        
        Args:
            texts: 输入文本列表
            
        Returns:
            embedding向量列表
        """
        if self.use_aliyun:
            return self._aliyun_embedding.embed_texts(texts)
        embeddings = self.model.encode(texts)
        return [emb.tolist() for emb in embeddings]
    
    def get_embedding_dimension(self) -> int:
        """获取embedding向量维度"""
        if self.use_aliyun:
            return self._aliyun_embedding.get_embedding_dimension()
        return self.model.get_sentence_embedding_dimension() if self.model else 0


def test_embedding():
    """测试embedding功能"""
    print("=" * 60)
    print("PostgreSQL Embedding 功能测试")
    print("=" * 60)
    
    try:
        embedding = PostgresEmbedding()
        
        print("\n" + "=" * 60)
        print("测试文本embedding")
        print("=" * 60)
        test_text = "Hello world"
        vector = embedding.embed_text(test_text)
        print(f"文本: {test_text}")
        print(f"向量维度: {len(vector)}")
        print(f"向量前5个值: {vector[:5]}")
        
        print("\n" + "=" * 60)
        print("测试批量embedding")
        print("=" * 60)
        test_texts = [
            "Hello world",
            "PostgreSQL is a database",
            "Embedding models convert text to vectors"
        ]
        vectors = embedding.embed_texts(test_texts)
        print(f"成功生成 {len(vectors)} 个向量")
        for i, vec in enumerate(vectors):
            print(f"  文本 {i+1}: {test_texts[i]}")
            print(f"    向量维度: {len(vec)}")
        
        print("\n" + "=" * 60)
        print("测试完成！")
        print("=" * 60)
        
        return True
        
    except Exception as e:
        print(f"测试失败: {e}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == "__main__":
    success = test_embedding()
    sys.exit(0 if success else 1)
