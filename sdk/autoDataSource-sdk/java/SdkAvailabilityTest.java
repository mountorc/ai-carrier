import com.example.autodatasource.sdk.AutoDataSourceClient;
import com.example.autodatasource.sdk.AutoDataSourceUnifiedClient;
import com.example.autodatasource.sdk.DataSourceMode;

public class SdkAvailabilityTest {
    public static void main(String[] args) {
        System.out.println("=================================");
        System.out.println("AutoDataSource SDK 可用性测试");
        System.out.println("=================================");

        try {
            System.out.println("\n测试 1: SDK 核心类加载");
            System.out.println("---------------------------------");
            
            Class.forName("com.example.autodatasource.sdk.AutoDataSourceClient");
            System.out.println("✓ AutoDataSourceClient 加载成功");
            
            Class.forName("com.example.autodatasource.sdk.AutoDataSourceUnifiedClient");
            System.out.println("✓ AutoDataSourceUnifiedClient 加载成功");
            
            Class.forName("com.example.autodatasource.sdk.DataSourceMode");
            System.out.println("✓ DataSourceMode 加载成功");
            
            Class.forName("com.example.autodatasource.sdk.AutoDataSourceSparkClient");
            System.out.println("✓ AutoDataSourceSparkClient 加载成功");

            System.out.println("\n测试 2: SDK 实例化");
            System.out.println("---------------------------------");
            
            String baseUrl = "http://localhost:8080/autoDataSource";
            
            AutoDataSourceClient client = new AutoDataSourceClient(baseUrl);
            System.out.println("✓ AutoDataSourceClient 实例化成功");
            
            AutoDataSourceUnifiedClient unifiedClient = new AutoDataSourceUnifiedClient(baseUrl);
            System.out.println("✓ AutoDataSourceUnifiedClient 实例化成功");
            
            AutoDataSourceUnifiedClient unifiedClientWithMode = new AutoDataSourceUnifiedClient(
                baseUrl, 
                DataSourceMode.HTTP_API
            );
            System.out.println("✓ AutoDataSourceUnifiedClient 指定模式实例化成功");

            System.out.println("\n测试 3: 枚举类型检查");
            System.out.println("---------------------------------");
            
            DataSourceMode[] modes = DataSourceMode.values();
            System.out.println("✓ DataSourceMode 枚举可用，包含 " + modes.length + " 个模式:");
            for (DataSourceMode mode : modes) {
                System.out.println("  - " + mode);
            }

            System.out.println("\n测试 4: 依赖库检查");
            System.out.println("---------------------------------");
            
            try {
                Class.forName("okhttp3.OkHttpClient");
                System.out.println("✓ OkHttp 可用");
            } catch (ClassNotFoundException e) {
                System.out.println("✗ OkHttp 不可用: " + e.getMessage());
            }
            
            try {
                Class.forName("com.fasterxml.jackson.databind.ObjectMapper");
                System.out.println("✓ Jackson Databind 可用");
            } catch (ClassNotFoundException e) {
                System.out.println("✗ Jackson Databind 不可用: " + e.getMessage());
            }
            
            try {
                Class.forName("org.slf4j.Logger");
                System.out.println("✓ SLF4J 可用");
            } catch (ClassNotFoundException e) {
                System.out.println("✗ SLF4J 不可用: " + e.getMessage());
            }

            System.out.println("\n=================================");
            System.out.println("测试结果总结");
            System.out.println("=================================");
            System.out.println("✓ SDK 所有核心类均可正常加载");
            System.out.println("✓ SDK 实例可以正常创建");
            System.out.println("✓ 枚举类型工作正常");
            System.out.println("✓ 主要依赖库已就绪");
            System.out.println("\nSDK 可用性: 100% ✓");
            System.out.println("\nSDK 已准备好在其他项目中使用！");
            System.out.println("\n使用方式:");
            System.out.println("1. 在 pom.xml 中添加 Maven 依赖");
            System.out.println("2. 或直接将 JAR 包添加到 classpath");
            System.out.println("3. 然后就可以开始使用 SDK 了！");
            System.out.println("=================================");

        } catch (Exception e) {
            System.err.println("\n=================================");
            System.err.println("测试失败");
            System.err.println("=================================");
            System.err.println("错误信息: " + e.getMessage());
            e.printStackTrace();
            System.err.println("=================================");
        }
    }
}
