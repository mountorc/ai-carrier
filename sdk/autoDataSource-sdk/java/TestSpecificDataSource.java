import com.example.autodatasource.sdk.AutoDataSourceClient;
import java.util.Map;

public class TestSpecificDataSource {
    public static void main(String[] args) {
        System.out.println("=================================");
        System.out.println("测试指定数据源");
        System.out.println("=================================");

        String baseUrl = "https://auto.xmzail.com/autoDataSource";
        String dataSourceId = "post4bc7-9a41-4332-93a1-a60c4d8a7e19";
        
        AutoDataSourceClient client = new AutoDataSourceClient(baseUrl);

        try {
            System.out.println("\n测试 1: 获取数据源列表");
            System.out.println("---------------------------------");
            Map<String, Object> dataSources = client.getDataSources();
            System.out.println("✓ 成功获取数据源列表");
            System.out.println("  响应: " + dataSources);

            System.out.println("\n测试 2: 获取该数据源的表列表");
            System.out.println("---------------------------------");
            try {
                Map<String, Object> tableList = client.getTableList(dataSourceId, null);
                System.out.println("✓ 成功获取表列表");
                System.out.println("  响应: " + tableList);
            } catch (Exception e) {
                System.out.println("✗ 获取表列表失败");
                System.out.println("  错误: " + e.getMessage());
            }

            System.out.println("\n测试 3: 预览该数据源的数据集");
            System.out.println("---------------------------------");
            try {
                Map<String, Object> preview = client.previewDataSetsByDataSourceId(dataSourceId);
                System.out.println("✓ 成功预览数据集");
                System.out.println("  响应: " + preview);
            } catch (Exception e) {
                System.out.println("✗ 预览数据集失败");
                System.out.println("  错误: " + e.getMessage());
            }

            System.out.println("\n测试 4: 执行简单SQL查询 (SELECT 1)");
            System.out.println("---------------------------------");
            try {
                Map<String, Object> queryResult = client.executeSqlQuery(dataSourceId, "SELECT 1");
                System.out.println("✓ 成功执行SQL查询");
                System.out.println("  响应: " + queryResult);
            } catch (Exception e) {
                System.out.println("✗ 执行SQL查询失败");
                System.out.println("  错误: " + e.getMessage());
            }

            System.out.println("\n=================================");
            System.out.println("测试完成");
            System.out.println("=================================");

        } catch (Exception e) {
            System.err.println("\n=================================");
            System.err.println("测试过程中发生错误");
            System.err.println("=================================");
            System.err.println("错误信息: " + e.getMessage());
            e.printStackTrace();
            System.err.println("=================================");
        }
    }
}
