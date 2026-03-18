import requests
import json

class AutoDataSourceClient:
    """AutoDataSource Python SDK 客户端"""
    
    def __init__(self, base_url, timeout=30):
        """初始化客户端
        
        Args:
            base_url: AutoDataSource API 基础 URL
            timeout: 请求超时时间（秒）
        """
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.session = requests.Session()
        self.session.headers.update({'Content-Type': 'application/json'})
    
    def get_data_sources(self):
        """获取数据源列表
        
        Returns:
            dict: 数据源列表响应
        
        Raises:
            requests.RequestException: 请求失败时抛出
        """
        url = f"{self.base_url}/api/data-sources/external/list"
        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"获取数据源列表失败: {str(e)}")
    
    def transform_sql(self, sql, source_type, target_type):
        """转换 SQL 语句
        
        Args:
            sql: SQL 语句
            source_type: 源数据库类型
            target_type: 目标数据库类型
            
        Returns:
            dict: SQL 转换响应
        
        Raises:
            requests.RequestException: 请求失败时抛出
        """
        url = f"{self.base_url}/api/sql/transform"
        data = {
            "sql": sql,
            "sourceType": source_type,
            "targetType": target_type
        }
        try:
            response = self.session.post(url, json=data, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"SQL 转换失败: {str(e)}")
    
    def get_data_items(self):
        """获取数据集列表
        
        Returns:
            dict: 数据集列表响应
        
        Raises:
            requests.RequestException: 请求失败时抛出
        """
        url = f"{self.base_url}/api/data-sets/list"
        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"获取数据集列表失败: {str(e)}")
    
    def get_data_by_uuid(self, uuid_auto_data):
        """根据 UUID 获取数据
        
        Args:
            uuid_auto_data: 数据集 UUID
            
        Returns:
            dict: 数据响应
        
        Raises:
            requests.RequestException: 请求失败时抛出
        """
        url = f"{self.base_url}/api/data-sets/data/{uuid_auto_data}"
        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"根据 UUID 获取数据失败: {str(e)}")
    
    def get_oss_files(self, prefix=None):
        """获取 OSS 文件列表
        
        Args:
            prefix: 目录前缀
            
        Returns:
            dict: OSS 文件列表响应
        
        Raises:
            requests.RequestException: 请求失败时抛出
        """
        url = f"{self.base_url}/api/oss/files"
        if prefix:
            url += f"?prefix={prefix}"
        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"获取 OSS 文件列表失败: {str(e)}")
    
    def get_public_docs_list(self):
        """获取公共文档列表
        
        Returns:
            dict: 文档列表响应
        
        Raises:
            requests.RequestException: 请求失败时抛出
        """
        url = f"{self.base_url}/docs-public/list"
        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"获取公共文档列表失败: {str(e)}")
    
    def get_public_doc_content(self, file_name):
        """获取公共文档内容
        
        Args:
            file_name: 文件名
            
        Returns:
            dict: 文档内容响应
        
        Raises:
            requests.RequestException: 请求失败时抛出
        """
        url = f"{self.base_url}/docs-public/docs?fileName={file_name}"
        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"获取公共文档内容失败: {str(e)}")
    
    def preview_data_items_by_data_source_id(self, data_source_id):
        """根据数据源ID预览数据集
        
        Args:
            data_source_id: 数据源ID
            
        Returns:
            dict: 数据集预览响应
        
        Raises:
            requests.RequestException: 请求失败时抛出
        """
        url = f"{self.base_url}/api/data-sets/preview/{data_source_id}"
        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"预览数据集失败: {str(e)}")
    
    def get_extract_records(self):
        """获取数据抽取记录
        
        Returns:
            dict: 数据抽取记录响应
        
        Raises:
            requests.RequestException: 请求失败时抛出
        """
        url = f"{self.base_url}/api/extract-records/list"
        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise Exception(f"获取数据抽取记录失败: {str(e)}")

__all__ = ['AutoDataSourceClient']
