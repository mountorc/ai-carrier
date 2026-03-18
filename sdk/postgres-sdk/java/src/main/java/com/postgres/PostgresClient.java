package com.postgres;

import com.fasterxml.jackson.databind.ObjectMapper;

import java.sql.*;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class PostgresClient {
    private Connection connection;
    private ObjectMapper objectMapper;
    private AliyunEmbedding aliyunEmbedding;
    private boolean useAliyun;

    public PostgresClient() {
        this(true, null, null, null);
    }
    
    public PostgresClient(boolean useAliyun) {
        this(useAliyun, null, null, null);
    }
    
    public PostgresClient(boolean useAliyun, String aliyunApiKey) {
        this(useAliyun, aliyunApiKey, null, null);
    }
    
    public PostgresClient(boolean useAliyun, String aliyunApiKey, String aliyunModel, String aliyunApiUrl) {
        this.objectMapper = new ObjectMapper();
        this.useAliyun = useAliyun;
        if (useAliyun) {
            this.aliyunEmbedding = new AliyunEmbedding(aliyunApiKey, aliyunModel, aliyunApiUrl);
        }
    }

    public void connectFromUrl(String url) throws SQLException {
        this.connection = DriverManager.getConnection(url);
        this.connection.setAutoCommit(true);
        System.out.println("Connected to PostgreSQL at " + url);
    }

    public void connect(String host, int port, String database, String user, String password) throws SQLException {
        String url = String.format("jdbc:postgresql://%s:%d/%s", host, port, database);
        this.connection = DriverManager.getConnection(url, user, password);
        this.connection.setAutoCommit(true);
        System.out.printf("Connected to PostgreSQL at %s:%d/%s%n", host, port, database);
    }

    public void disconnect() {
        if (connection != null) {
            try {
                connection.close();
                System.out.println("Disconnected from PostgreSQL");
            } catch (SQLException e) {
                e.printStackTrace();
            }
        }
    }

    public void createCollection(String collectionName, int dimension, String indexType) throws SQLException {
        String createTableSql = String.format("""
            CREATE TABLE IF NOT EXISTS %s (
                id SERIAL PRIMARY KEY,
                embedding vector(%d),
                metadata jsonb
            )
        """, collectionName, dimension);

        try (Statement stmt = connection.createStatement()) {
            stmt.execute(createTableSql);
        }

        if (indexType != null && !indexType.isEmpty()) {
            String createIndexSql;
            switch (indexType.toUpperCase()) {
                case "IVFFLAT":
                    createIndexSql = String.format("""
                        CREATE INDEX IF NOT EXISTS %s_embedding_idx
                        ON %s USING ivfflat (embedding vector_l2_ops)
                        WITH (lists = 128)
                    """, collectionName, collectionName);
                    break;
                case "HNSW":
                    createIndexSql = String.format("""
                        CREATE INDEX IF NOT EXISTS %s_embedding_idx
                        ON %s USING hnsw (embedding vector_l2_ops)
                        WITH (m = 16, ef_construction = 64)
                    """, collectionName, collectionName);
                    break;
                default:
                    throw new IllegalArgumentException("Unsupported index type: " + indexType);
            }

            try (Statement stmt = connection.createStatement()) {
                stmt.execute(createIndexSql);
            }
        }

        System.out.printf("Collection %s created with dimension %d%n", collectionName, dimension);
    }

    public void dropCollection(String collectionName) throws SQLException {
        String dropTableSql = String.format("DROP TABLE IF EXISTS %s", collectionName);
        try (Statement stmt = connection.createStatement()) {
            stmt.execute(dropTableSql);
        }
        System.out.printf("Collection %s dropped%n", collectionName);
    }

    public List<String> listCollections() throws SQLException {
        List<String> collections = new ArrayList<>();
        String sql = "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'";

        try (Statement stmt = connection.createStatement();
             ResultSet rs = stmt.executeQuery(sql)) {
            while (rs.next()) {
                collections.add(rs.getString("table_name"));
            }
        }
        return collections;
    }

    public void insertVectors(String collectionName, List<double[]> vectors, List<Map<String, Object>> metadata) throws Exception {
        if (metadata == null) {
            metadata = new ArrayList<>();
            for (int i = 0; i < vectors.size(); i++) {
                metadata.add(new HashMap<>());
            }
        }

        String insertSql = String.format("INSERT INTO %s (embedding, metadata) VALUES (?::vector, ?::jsonb)", collectionName);

        try (PreparedStatement pstmt = connection.prepareStatement(insertSql)) {
            for (int i = 0; i < vectors.size(); i++) {
                String vectorStr = formatVector(vectors.get(i));
                String metadataJson = objectMapper.writeValueAsString(metadata.get(i));

                pstmt.setString(1, vectorStr);
                pstmt.setString(2, metadataJson);
                pstmt.addBatch();
            }
            pstmt.executeBatch();
        }

        System.out.printf("Inserted %d vectors into %s%n", vectors.size(), collectionName);
    }

    public static class SearchResult {
        public int id;
        public double score;
        public Map<String, Object> metadata;
    }

    public List<SearchResult> searchVectors(String collectionName, double[] queryVector, int topK, String filterCondition) throws Exception {
        String vectorStr = formatVector(queryVector);
        List<SearchResult> results = new ArrayList<>();

        String sql;
        if (filterCondition != null && !filterCondition.isEmpty()) {
            sql = String.format("""
                SELECT id, embedding <-> ?::vector as distance, metadata
                FROM %s
                WHERE %s
                ORDER BY distance
                LIMIT ?
            """, collectionName, filterCondition);
        } else {
            sql = String.format("""
                SELECT id, embedding <-> ?::vector as distance, metadata
                FROM %s
                ORDER BY distance
                LIMIT ?
            """, collectionName);
        }

        try (PreparedStatement pstmt = connection.prepareStatement(sql)) {
            pstmt.setString(1, vectorStr);
            pstmt.setInt(2, topK);

            try (ResultSet rs = pstmt.executeQuery()) {
                while (rs.next()) {
                    SearchResult result = new SearchResult();
                    result.id = rs.getInt("id");
                    result.score = rs.getDouble("distance");
                    String metadataJson = rs.getString("metadata");
                    result.metadata = objectMapper.readValue(metadataJson, Map.class);
                    results.add(result);
                }
            }
        }

        return results;
    }

    public List<Map<String, Object>> executeSql(String sql, Object... params) throws SQLException {
        List<Map<String, Object>> results = new ArrayList<>();

        try (PreparedStatement pstmt = connection.prepareStatement(sql)) {
            for (int i = 0; i < params.length; i++) {
                pstmt.setObject(i + 1, params[i]);
            }

            try (ResultSet rs = pstmt.executeQuery()) {
                ResultSetMetaData md = rs.getMetaData();
                int columns = md.getColumnCount();

                while (rs.next()) {
                    Map<String, Object> row = new HashMap<>();
                    for (int i = 1; i <= columns; i++) {
                        row.put(md.getColumnName(i), rs.getObject(i));
                    }
                    results.add(row);
                }
            }
        }

        return results;
    }

    public List<Map<String, Object>> getTableFields(String tableSchema, String tableName) throws SQLException {
        String sql = """
            SELECT column_name, data_type, character_maximum_length
            FROM information_schema.columns
            WHERE table_schema = ? AND table_name = ?
            ORDER BY ordinal_position
        """;
        return executeSql(sql, tableSchema, tableName);
    }

    public List<String> getTableList(String tableSchema) throws SQLException {
        if (tableSchema == null || tableSchema.isEmpty()) {
            tableSchema = "public";
        }

        String sql = """
            SELECT table_name
            FROM information_schema.tables
            WHERE table_schema = ?
            ORDER BY table_name
        """;
        List<Map<String, Object>> results = executeSql(sql, tableSchema);

        List<String> tables = new ArrayList<>();
        for (Map<String, Object> row : results) {
            tables.add((String) row.get("table_name"));
        }
        return tables;
    }

    public void beginTransaction() throws SQLException {
        connection.setAutoCommit(false);
    }

    public void commit() throws SQLException {
        connection.commit();
        connection.setAutoCommit(true);
    }

    public void rollback() throws SQLException {
        connection.rollback();
        connection.setAutoCommit(true);
    }

    public double[] getEmbedding(String text) throws Exception {
        if (useAliyun) {
            return aliyunEmbedding.embedText(text);
        }
        throw new UnsupportedOperationException("Local embedding not implemented yet. Use useAliyun=true.");
    }
    
    public List<SearchResult> searchByText(String tableName, String queryText, String vectorColumn, 
                                            int topK, String distanceType, List<String> selectColumns) throws Exception {
        double[] queryVector = getEmbedding(queryText);
        return searchJsonbVectors(tableName, queryVector, vectorColumn, topK, distanceType, selectColumns);
    }
    
    public List<Map<String, Object>> searchJsonbVectors(String tableName, double[] queryVector, 
                                                         String vectorColumn, int topK, String distanceType, 
                                                         List<String> selectColumns) throws Exception {
        String columnsSql = selectColumns != null && !selectColumns.isEmpty() 
                ? String.join(", ", selectColumns) 
                : "*";
        
        String sql = String.format("SELECT %s, %s FROM %s", columnsSql, vectorColumn, tableName);
        
        List<Map<String, Object>> results = new ArrayList<>();
        
        try (Statement stmt = connection.createStatement();
             ResultSet rs = stmt.executeQuery(sql)) {
            
            while (rs.next()) {
                Map<String, Object> row = new HashMap<>();
                ResultSetMetaData md = rs.getMetaData();
                int columns = md.getColumnCount();
                
                double[] vec = null;
                for (int i = 1; i <= columns; i++) {
                    String colName = md.getColumnName(i);
                    if (colName.equals(vectorColumn)) {
                        Object vecData = rs.getObject(i);
                        if (vecData instanceof String) {
                            vec = objectMapper.readValue((String) vecData, double[].class);
                        }
                    } else {
                        row.put(colName, rs.getObject(i));
                    }
                }
                
                if (vec != null) {
                    double distance = calculateDistance(queryVector, vec, distanceType);
                    row.put("distance", distance);
                    results.add(row);
                }
            }
        }
        
        results.sort((a, b) -> Double.compare((Double) a.get("distance"), (Double) b.get("distance")));
        return results.subList(0, Math.min(topK, results.size()));
    }
    
    private double calculateDistance(double[] v1, double[] v2, String distanceType) {
        if ("l2".equals(distanceType)) {
            return l2Distance(v1, v2);
        } else if ("cosine".equals(distanceType)) {
            return 1 - cosineSimilarity(v1, v2);
        } else if ("inner_product".equals(distanceType)) {
            return -innerProduct(v1, v2);
        }
        throw new IllegalArgumentException("Unsupported distance type: " + distanceType);
    }
    
    private double l2Distance(double[] v1, double[] v2) {
        double sum = 0;
        for (int i = 0; i < v1.length; i++) {
            sum += (v1[i] - v2[i]) * (v1[i] - v2[i]);
        }
        return Math.sqrt(sum);
    }
    
    private double cosineSimilarity(double[] v1, double[] v2) {
        double dotProduct = 0, norm1 = 0, norm2 = 0;
        for (int i = 0; i < v1.length; i++) {
            dotProduct += v1[i] * v2[i];
            norm1 += v1[i] * v1[i];
            norm2 += v2[i] * v2[i];
        }
        if (norm1 == 0 || norm2 == 0) return 0;
        return dotProduct / (Math.sqrt(norm1) * Math.sqrt(norm2));
    }
    
    private double innerProduct(double[] v1, double[] v2) {
        double sum = 0;
        for (int i = 0; i < v1.length; i++) {
            sum += v1[i] * v2[i];
        }
        return sum;
    }
    
    private String formatVector(double[] vector) {
        StringBuilder sb = new StringBuilder("[");
        for (int i = 0; i < vector.length; i++) {
            if (i > 0) {
                sb.append(", ");
            }
            sb.append(vector[i]);
        }
        sb.append("]");
        return sb.toString();
    }
}
