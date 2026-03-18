from setuptools import setup, find_packages

setup(
    name="autodatasource-sdk-python",
    version="1.1.0",
    packages=find_packages(),
    install_requires=[
        "requests>=2.28.0"
    ],
    author="XMZ",
    author_email="example@example.com",
    description="AutoDataSource Python SDK",
    long_description="""AutoDataSource Python SDK 提供了与 AutoDataSource API 交互的功能，包括获取数据源列表、转换 SQL 语句、管理数据集等。""",
    url="https://github.com/example/autodatasource-sdk-python",
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    python_requires='>=3.6',
)
