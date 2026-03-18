package com.postgres;

import com.fasterxml.jackson.databind.ObjectMapper;
import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class AliyunEmbedding {
    public static final String DEFAULT_API_URL = "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings";
    public static final String DEFAULT_MODEL = "text-embedding-v4";
    public static final String DEFAULT_API_KEY = "sk-403ca84daa9740df82ce0a1737ceccdf";
    
    private final String apiKey;
    private final String model;
    private final String apiUrl;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;
    private Integer dimension;
    
    public AliyunEmbedding() {
        this(DEFAULT_API_KEY, DEFAULT_MODEL, DEFAULT_API_URL);
    }
    
    public AliyunEmbedding(String apiKey) {
        this(apiKey, DEFAULT_MODEL, DEFAULT_API_URL);
    }
    
    public AliyunEmbedding(String apiKey, String model) {
        this(apiKey, model, DEFAULT_API_URL);
    }
    
    public AliyunEmbedding(String apiKey, String model, String apiUrl) {
        this.apiKey = apiKey != null ? apiKey : DEFAULT_API_KEY;
        this.model = model != null ? model : DEFAULT_MODEL;
        this.apiUrl = apiUrl != null ? apiUrl : DEFAULT_API_URL;
        this.httpClient = HttpClient.newHttpClient();
        this.objectMapper = new ObjectMapper();
        System.out.println("初始化阿里云百炼Embedding: 模型=" + this.model);
    }
    
    public double[] embedText(String text) throws IOException, InterruptedException {
        Map<String, Object> requestBody = new HashMap<>();
        requestBody.put("model", model);
        requestBody.put("input", text);
        
        String jsonBody = objectMapper.writeValueAsString(requestBody);
        
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(apiUrl))
                .header("Authorization", "Bearer " + apiKey)
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(jsonBody))
                .build();
        
        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        
        if (response.statusCode() != 200) {
            throw new RuntimeException("阿里云API调用失败: " + response.statusCode() + ", " + response.body());
        }
        
        Map<String, Object> result = objectMapper.readValue(response.body(), Map.class);
        
        @SuppressWarnings("unchecked")
        List<Map<String, Object>> data = (List<Map<String, Object>>) result.get("data");
        
        if (data == null || data.isEmpty()) {
            throw new RuntimeException("阿里云API返回格式错误: " + result);
        }
        
        @SuppressWarnings("unchecked")
        List<Double> embedding = (List<Double>) data.get(0).get("embedding");
        
        double[] vector = new double[embedding.size()];
        for (int i = 0; i < embedding.size(); i++) {
            vector[i] = embedding.get(i);
        }
        
        if (this.dimension == null) {
            this.dimension = vector.length;
            System.out.println("模型维度: " + this.dimension);
        }
        
        return vector;
    }
    
    public double[][] embedTexts(List<String> texts) throws IOException, InterruptedException {
        double[][] vectors = new double[texts.size()][];
        for (int i = 0; i < texts.size(); i++) {
            vectors[i] = embedText(texts.get(i));
        }
        return vectors;
    }
    
    public int getEmbeddingDimension() throws IOException, InterruptedException {
        if (this.dimension == null) {
            embedText("test");
        }
        return this.dimension;
    }
}
