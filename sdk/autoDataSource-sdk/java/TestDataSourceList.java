import com.example.autodatasource.sdk.AutoDataSourceClient;
import java.util.Map;

public class TestDataSourceList {
    public static void main(String[] args) {
        // 初始化客户端
        String baseUrl = "http://localhost:8080/autoDataSource";
        AutoDataSourceClient client = new AutoDataSourceClient(baseUrl);

        try {
            // 获取数据源列表
            Map<String, Object> dataSourcesResponse = client.getDataSources();
            System.out.println("数据源列表响应: " + dataSourcesResponse);
            System.out.println("成功获取数据源列表！");
        } catch (Exception e) {
            System.err.println("获取数据源列表失败: " + e.getMessage());
            e.printStackTrace();
        }
    }
}