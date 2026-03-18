#!/usr/bin/env python3
"""AutoDataSource Python SDK 使用示例"""

from autodatasource import AutoDataSourceClient

def main():
    """主函数"""
    # 初始化客户端
    client = AutoDataSourceClient("http://localhost:8080/autoDataSource")
    
    print("=== 示例 1: 获取数据源列表 ===")
    try:
        response = client.get_data_sources()
        print("响应:", response)
    except Exception as e:
        print("错误:", str(e))
    print()
    
    print("=== 示例 2: 转换 SQL 语句 ===")
    try:
        sql = "SELECT * FROM users LIMIT 10"
        response = client.transform_sql(sql, "mysql", "oracle")
        print("响应:", response)
    except Exception as e:
        print("错误:", str(e))
    print()
    
    print("=== 示例 3: 获取数据集列表 ===")
    try:
        response = client.get_data_items()
        print("响应:", response)
    except Exception as e:
        print("错误:", str(e))
    print()
    
    print("=== 示例 4: 根据 UUID 获取数据 ===")
    try:
        uuid = "e38ff77ff5b84985aff4cb97ebb87409"
        response = client.get_data_by_uuid(uuid)
        print("响应:", response)
    except Exception as e:
        print("错误:", str(e))
    print()
    
    print("=== 示例 5: 获取 OSS 文件列表 ===")
    try:
        response = client.get_oss_files()
        print("响应:", response)
    except Exception as e:
        print("错误:", str(e))
    print()
    
    print("=== 示例 6: 获取公共文档列表 ===")
    try:
        response = client.get_public_docs_list()
        print("响应:", response)
    except Exception as e:
        print("错误:", str(e))
    print()
    
    print("=== 示例 7: 获取公共文档内容 ===")
    try:
        response = client.get_public_doc_content("README.md")
        print("响应:", response)
    except Exception as e:
        print("错误:", str(e))
    print()
    
    print("=== 示例 8: 根据数据源ID预览数据集 ===")
    try:
        data_source_id = "90a49r2l313l4243robp5lbqf678vfn10719"
        response = client.preview_data_items_by_data_source_id(data_source_id)
        print("响应:", response)
    except Exception as e:
        print("错误:", str(e))
    print()

if __name__ == "__main__":
    main()
