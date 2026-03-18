package com.postgres.example;

import com.postgres.PostgresClient;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class Main {
    public static void main(String[] args) {
        PostgresClient client = new PostgresClient();

        try {
            String dbUrl = "jdbc:postgresql://121.43.142.153:5432/carrier?user=carrier&password=GNerfiSP4dpZjwcJ";
            System.out.println("Connecting to database...");
            client.connectFromUrl(dbUrl);

            System.out.println("\n=== List collections ===");
            List<String> collections = client.listCollections();
            System.out.println("Collections: " + collections);

            String collectionName = "test_vectors_java";
            int dimension = 128;

            System.out.printf("\n=== Create collection %s with dimension %d ===%n", collectionName, dimension);
            client.createCollection(collectionName, dimension, "IVFFLAT");

            System.out.println("\n=== Insert vectors ===");
            List<double[]> vectors = new ArrayList<>();
            for (int i = 0; i < 3; i++) {
                double[] vec = new double[dimension];
                for (int j = 0; j < dimension; j++) {
                    vec[j] = (i + 1) * 0.1 * j;
                }
                vectors.add(vec);
            }

            List<Map<String, Object>> metadata = new ArrayList<>();
            Map<String, Object> meta1 = new HashMap<>();
            meta1.put("text", "Hello world");
            meta1.put("category", "test");
            metadata.add(meta1);

            Map<String, Object> meta2 = new HashMap<>();
            meta2.put("text", "Java SDK");
            meta2.put("category", "sdk");
            metadata.add(meta2);

            Map<String, Object> meta3 = new HashMap<>();
            meta3.put("text", "PostgreSQL vector");
            meta3.put("category", "database");
            metadata.add(meta3);

            client.insertVectors(collectionName, vectors, metadata);

            System.out.println("\n=== Search vectors ===");
            double[] queryVector = new double[dimension];
            for (int j = 0; j < dimension; j++) {
                queryVector[j] = 0.15 * j;
            }

            List<PostgresClient.SearchResult> results = client.searchVectors(collectionName, queryVector, 3, "");
            for (PostgresClient.SearchResult result : results) {
                System.out.printf("ID: %d, Score: %.4f, Metadata: %s%n", result.id, result.score, result.metadata);
            }

            System.out.println("\n=== Get table list ===");
            List<String> tables = client.getTableList("");
            System.out.println("Tables: " + tables);

            if (!tables.isEmpty()) {
                System.out.printf("\n=== Get fields for table %s ===%n", tables.get(0));
                List<Map<String, Object>> fields = client.getTableFields("public", tables.get(0));
                System.out.println("Fields: " + fields);
            }

            System.out.println("\n=== Drop collection ===");
            client.dropCollection(collectionName);

        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
            e.printStackTrace();
        } finally {
            client.disconnect();
        }
    }
}
