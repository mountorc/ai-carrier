#!/usr/bin/env python3
import os
import re

def update_imports_in_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original_content = content
    
    # 更新导入路径
    # github.com/example/postgres-sdk-go/postgres -> github.com/example/dataset-sdk/postgres
    content = re.sub(
        r'github\.com/example/postgres-sdk-go/postgres',
        'github.com/example/dataset-sdk/postgres',
        content
    )
    
    # github.com/example/datasource -> github.com/example/dataset-sdk/datasource
    content = re.sub(
        r'github\.com/example/datasource',
        'github.com/example/dataset-sdk/datasource',
        content
    )
    
    # github.com/example/dataset-go/database -> github.com/example/dataset-sdk/database
    content = re.sub(
        r'github\.com/example/dataset-go/database',
        'github.com/example/dataset-sdk/database',
        content
    )
    
    # github.com/example/dataset-go/scheduler -> github.com/example/dataset-sdk/scheduler
    content = re.sub(
        r'github\.com/example/dataset-go/scheduler',
        'github.com/example/dataset-sdk/scheduler',
        content
    )
    
    # github.com/example/dataset-go/server -> github.com/example/dataset-sdk/server
    content = re.sub(
        r'github\.com/example/dataset-go/server',
        'github.com/example/dataset-sdk/server',
        content
    )
    
    if content != original_content:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"Updated: {file_path}")
        return True
    return False

def main():
    sdk_dir = os.path.dirname(os.path.abspath(__file__))
    
    for root, dirs, files in os.walk(sdk_dir):
        for file in files:
            if file.endswith('.go'):
                file_path = os.path.join(root, file)
                update_imports_in_file(file_path)
    
    print("Done!")

if __name__ == '__main__':
    main()
