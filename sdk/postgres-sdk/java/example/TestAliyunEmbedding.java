package example;

import com.postgres.AliyunEmbedding;
import com.postgres.PostgresClient;

import java.util.Arrays;
import java.util.List;

public class TestAliyunEmbedding {
    public static void main(String[] args) throws Exception {
        System.out.println("=".repeat(60));
        System.out.println("测试 Java SDK 阿里云百炼 Embedding");
        System.out.println("=".repeat(60));
        
        // 1. 测试直接使用 AliyunEmbedding
        System.out.println("\n1. 测试直接使用 AliyunEmbedding...");
        AliyunEmbedding embedding = new AliyunEmbedding();
        System.out.println("✓ AliyunEmbedding 初始化成功!");
        
        System.out.println("\n" + "=".repeat(60));
        System.out.println("2. 测试单个文本embedding");
        System.out.println("=".repeat(60));
        
        List<String> testTexts = Arrays.asList(
            "Hello world",
            "你好世界",
            "衣服的质量杠杠的",
            "数据处理"
        );
        
        for (String text : testTexts) {
            System.out.println("\n文本: " + text);
            double[] vector = embedding.embedText(text);
            System.out.println("  向量维度: " + vector.length);
            System.out.print("  前5个值: [");
            for (int i = 0; i < 5 && i < vector.length; i++) {
                if (i > 0) System.out.print(", ");
                System.out.printf("%.4f", vector[i]);
            }
            System.out.println("]");
        }
        
        System.out.println("\n" + "=".repeat(60));
        System.out.println("3. 测试 PostgresClient 使用阿里云");
        System.out.println("=".repeat(60));
        
        PostgresClient client = new PostgresClient(true);
        System.out.println("✓ PostgresClient (useAliyun=true) 初始化成功!");
        
        String text = "机器学习";
        System.out.println("\n文本: " + text);
        double[] vector = client.getEmbedding(text);
        System.out.println("  向量维度: " + vector.length);
        System.out.print("  前5个值: [");
        for (int i = 0; i < 5 && i < vector.length; i++) {
            if (i > 0) System.out.print(", ");
            System.out.printf("%.4f", vector[i]);
        }
        System.out.println("]");
        
        System.out.println("\n" + "=".repeat(60));
        System.out.println("✓ 所有测试通过!");
        System.out.println("=".repeat(60));
        System.out.println("\n使用说明:");
        System.out.println("  // 直接使用 AliyunEmbedding");
        System.out.println("  AliyunEmbedding embedding = new AliyunEmbedding();");
        System.out.println("  double[] vector = embedding.embedText(\"你的文本\");");
        System.out.println();
        System.out.println("  // 通过 PostgresClient 使用");
        System.out.println("  PostgresClient client = new PostgresClient(true);");
        System.out.println("  double[] vector = client.getEmbedding(\"你的文本\");");
    }
}
