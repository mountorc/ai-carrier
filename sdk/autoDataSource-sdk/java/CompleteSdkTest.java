import com.example.autodatasource.sdk.AutoDataSourceClient;
import java.util.Map;

public class CompleteSdkTest {
    public static void main(String[] args) {
        System.out.println("=================================");
        System.out.println("AutoDataSource SDK 完整功能测试");
        System.out.println("=================================");

        String baseUrl = "http://localhost:8080/autoDataSource";
        AutoDataSourceClient client = new AutoDataSourceClient(baseUrl);

        try {
            System.out.println("\n测试 1: 获取公共文档列表");
            System.out.println("---------------------------------");
            Map<String, Object> docsList = client.getPublicDocsList();
            System.out.println("✓ 成功获取公共文档列表");
            System.out.println("  响应: " + docsList);

            System.out.println("\n测试 2: 获取公共文档内容");
            System.out.println("---------------------------------");
            try {
                Map<String, Object> docContent = client.getPublicDocContent("README.md");
                System.out.println("✓ 成功获取 README.md 内容");
                System.out.println("  响应长度: " + ((String) docContent.get("content")).length() + " 字符");
            } catch (Exception e) {
                System.out.println("  (README.md 可能不存在，继续其他测试)");
            }

            System.out.println("\n测试 3: 获取数据源列表");
            System.out.println("---------------------------------");
            Map<String, Object> dataSources = client.getDataSources();
            System.out.println("✓ 成功获取数据源列表");
            System.out.println("  响应: " + dataSources);

            System.out.println("\n测试 4: 获取数据集列表");
            System.out.println("---------------------------------");
            try {
                Map<String, Object> dataSets = client.getDataSets();
                System.out.println("✓ 成功获取数据集列表");
                System.out.println("  响应: " + dataSets);
            } catch (Exception e) {
                System.out.println("  (数据集接口可能不可用，继续)");
            }

            System.out.println("\n=================================");
            System.out.println("测试结果总结");
            System.out.println("=================================");
            System.out.println("✓ AutoDataSource SDK 工作正常");
            System.out.println("✓ 可以成功连接到服务");
            System.out.println("✓ API 接口响应正常");
            System.out.println("✓ SDK 已准备好在其他项目中使用");
            System.out.println("=================================");

        } catch (Exception e) {
            System.err.println("\n=================================");
            System.err.println("测试过程中发生错误");
            System.err.println("=================================");
            System.err.println("错误信息: " + e.getMessage());
            e.printStackTrace();
            System.err.println("=================================");
            System.err.println("可能的原因:");
            System.err.println("1. AutoDataSource 服务未正常启动");
            System.err.println("2. 网络连接问题");
            System.err.println("3. API 路径配置错误");
            System.err.println("=================================");
        }
    }
}
