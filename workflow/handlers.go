// services/mountcore/workflow/handlers.go

package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/trae/autoFlow/common/httpclient"
	"github.com/trae/autoFlow/common/plugin"
	"github.com/trae/autoFlow/mounts/workflow"
)

// 辅助函数：获取项目根目录下的正确文件路径
func getProjectFilePath(relativePath string) string {
	// 先尝试从当前目录开始查找
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("获取当前工作目录失败: %v", err)
		return relativePath
	}

	// 可能的路径候选
	candidates := []string{
		filepath.Join(cwd, relativePath),
		filepath.Join(cwd, "..", relativePath),
		filepath.Join(cwd, "..", "..", relativePath),
		"./" + relativePath,
		"../" + relativePath,
		"../../" + relativePath,
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			log.Printf("找到文件: %s", candidate)
			return candidate
		}
	}

	log.Printf("警告: 未能找到文件 %s，使用默认路径: ./%s", relativePath, relativePath)
	return "./" + relativePath
}

// 辅助函数：获取正确的数据文件路径
func getDataFilePath(filename string) string {
	return getProjectFilePath("data/" + filename)
}

// FlowExecuteRequest 定义流程执行请求结构
type FlowExecuteRequest struct {
	Flow *workflow.FlowDefinition `json:"flow"` // 流程定义
}

// FlowExecuteResponse 定义流程执行响应结构
type FlowExecuteResponse struct {
	Success       bool                     `json:"success"`                  // 是否成功
	InstanceID    string                   `json:"instance_id,omitempty"`    // 流程实例ID
	Logs          []map[string]interface{} `json:"logs"`                     // 执行日志
	FinalOutput   interface{}              `json:"final_output,omitempty"`   // 最终输出
	ExecutedNodes int                      `json:"executed_nodes,omitempty"` // 执行的节点数
	Error         string                   `json:"error,omitempty"`          // 错误信息
}

// GetFlowsResponse 定义获取工作流列表响应结构
type GetFlowsResponse struct {
	Success bool                      `json:"success"`         // 是否成功
	Flows   []workflow.FlowDefinition `json:"flows,omitempty"` // 工作流列表
	Error   string                    `json:"error,omitempty"` // 错误信息
}

// PluginExecuteRequest 定义插件执行请求结构
type PluginExecuteRequest struct {
	PluginID   string                 `json:"plugin_id"`         // 插件ID
	Parameters map[string]interface{} `json:"parameters"`        // 插件参数
	Timeout    int                    `json:"timeout,omitempty"` // 超时时间
}

// PluginExecuteResponse 定义插件执行响应结构
type PluginExecuteResponse struct {
	Success bool                   `json:"success"`
	Outputs map[string]interface{} `json:"outputs,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// ExecutionRecord 工作流执行记录
type ExecutionRecord struct {
	ID            string                   `json:"id"`             // 执行记录ID
	FlowID        string                   `json:"flow_id"`        // 工作流ID
	FlowName      string                   `json:"flow_name"`      // 工作流名称
	StartTime     string                   `json:"start_time"`     // 开始时间
	EndTime       string                   `json:"end_time"`       // 结束时间
	Status        string                   `json:"status"`         // 执行状态（success/error）
	DurationMs    int                      `json:"duration_ms"`    // 执行时长（毫秒）
	Input         map[string]interface{}   `json:"input"`          // 输入参数
	Output        map[string]interface{}   `json:"output"`         // 输出结果
	NodesExecuted int                      `json:"nodes_executed"` // 执行的节点数
	Errors        []ExecutionError         `json:"errors"`         // 错误信息
	Logs          []map[string]interface{} `json:"logs"`           // 执行日志
}

// ExecutionError 执行错误信息
type ExecutionError struct {
	NodeID  string `json:"node_id"` // 节点ID
	Message string `json:"message"` // 错误消息
}

// ExecutionRecords 执行记录集合
type ExecutionRecords struct {
	Records []ExecutionRecord              `json:"records"` // 执行记录列表
	Index   map[string]map[string][]string `json:"index"`   // 索引映射
}

// LoadExecutionRecords 加载执行记录
func LoadExecutionRecords(filePath string) (*ExecutionRecords, error) {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("创建目录失败: %v", err)
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，返回空记录
		return &ExecutionRecords{
			Records: []ExecutionRecord{},
			Index: map[string]map[string][]string{
				"by_flow_id": {},
				"by_status":  {},
				"by_date":    {},
			},
		}, nil
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 解析JSON
	var records ExecutionRecords
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &records, nil
}

// SaveExecutionRecords 保存执行记录
func SaveExecutionRecords(filePath string, records *ExecutionRecords) error {
	// 序列化JSON
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// AddExecutionRecord 添加执行记录
func AddExecutionRecord(filePath string, record ExecutionRecord) error {
	// 加载现有记录
	records, err := LoadExecutionRecords(filePath)
	if err != nil {
		return err
	}

	// 生成记录ID（如果未提供）
	if record.ID == "" {
		record.ID = fmt.Sprintf("exec_%d", time.Now().UnixMilli())
	}

	// 添加记录
	records.Records = append(records.Records, record)

	// 更新索引
	updateExecutionRecordIndex(records, record)

	// 保存记录
	return SaveExecutionRecords(filePath, records)
}

// 更新索引
func updateExecutionRecordIndex(records *ExecutionRecords, record ExecutionRecord) {
	// 确保索引映射存在
	if records.Index == nil {
		records.Index = map[string]map[string][]string{
			"by_flow_id": {},
			"by_status":  {},
			"by_date":    {},
		}
	}

	// 按流程ID索引
	if records.Index["by_flow_id"] == nil {
		records.Index["by_flow_id"] = map[string][]string{}
	}
	records.Index["by_flow_id"][record.FlowID] = append(records.Index["by_flow_id"][record.FlowID], record.ID)

	// 按状态索引
	if records.Index["by_status"] == nil {
		records.Index["by_status"] = map[string][]string{}
	}
	records.Index["by_status"][record.Status] = append(records.Index["by_status"][record.Status], record.ID)

	// 按日期索引
	if records.Index["by_date"] == nil {
		records.Index["by_date"] = map[string][]string{}
	}
	// 提取日期部分（YYYY-MM-DD）
	startTime, err := time.Parse(time.RFC3339, record.StartTime)
	if err == nil {
		dateKey := startTime.Format("2006-01-02")
		records.Index["by_date"][dateKey] = append(records.Index["by_date"][dateKey], record.ID)
	}
}

// GetExecutionRecordsByFlowID 按流程ID查询执行记录
func GetExecutionRecordsByFlowID(filePath string, flowID string) ([]ExecutionRecord, error) {
	// 加载记录
	records, err := LoadExecutionRecords(filePath)
	if err != nil {
		return nil, err
	}

	// 查找索引
	recordIDs, ok := records.Index["by_flow_id"][flowID]
	if !ok {
		return []ExecutionRecord{}, nil
	}

	// 查找对应的记录
	result := make([]ExecutionRecord, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		for _, record := range records.Records {
			if record.ID == recordID {
				result = append(result, record)
				break
			}
		}
	}

	// 按时间倒序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime > result[j].StartTime
	})

	return result, nil
}

// GetExecutionRecordsByStatus 按状态查询执行记录
func GetExecutionRecordsByStatus(filePath string, status string) ([]ExecutionRecord, error) {
	// 加载记录
	records, err := LoadExecutionRecords(filePath)
	if err != nil {
		return nil, err
	}

	// 查找索引
	recordIDs, ok := records.Index["by_status"][status]
	if !ok {
		return []ExecutionRecord{}, nil
	}

	// 查找对应的记录
	result := make([]ExecutionRecord, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		for _, record := range records.Records {
			if record.ID == recordID {
				result = append(result, record)
				break
			}
		}
	}

	// 按时间倒序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime > result[j].StartTime
	})

	return result, nil
}

// GetExecutionRecordsByDate 按日期查询执行记录
func GetExecutionRecordsByDate(filePath string, date string) ([]ExecutionRecord, error) {
	// 加载记录
	records, err := LoadExecutionRecords(filePath)
	if err != nil {
		return nil, err
	}

	// 查找索引
	recordIDs, ok := records.Index["by_date"][date]
	if !ok {
		return []ExecutionRecord{}, nil
	}

	// 查找对应的记录
	result := make([]ExecutionRecord, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		for _, record := range records.Records {
			if record.ID == recordID {
				result = append(result, record)
				break
			}
		}
	}

	// 按时间倒序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime > result[j].StartTime
	})

	return result, nil
}

// GetAllExecutionRecords 获取所有执行记录
func GetAllExecutionRecords(filePath string) ([]ExecutionRecord, error) {
	// 加载记录
	records, err := LoadExecutionRecords(filePath)
	if err != nil {
		return nil, err
	}

	// 按时间倒序排序
	result := make([]ExecutionRecord, len(records.Records))
	copy(result, records.Records)
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime > result[j].StartTime
	})

	return result, nil
}

// GetExecutionRecordByID 根据ID获取执行记录
func GetExecutionRecordByID(filePath string, id string) (*ExecutionRecord, error) {
	// 加载记录
	records, err := LoadExecutionRecords(filePath)
	if err != nil {
		return nil, err
	}

	// 查找记录
	for _, record := range records.Records {
		if record.ID == id {
			return &record, nil
		}
	}

	return nil, fmt.Errorf("记录不存在: %s", id)
}

// DeleteExecutionRecord 删除执行记录
func DeleteExecutionRecord(filePath string, id string) error {
	// 加载记录
	records, err := LoadExecutionRecords(filePath)
	if err != nil {
		return err
	}

	// 查找并删除记录
	newRecords := make([]ExecutionRecord, 0, len(records.Records))
	deleted := false
	for _, record := range records.Records {
		if record.ID != id {
			newRecords = append(newRecords, record)
		} else {
			deleted = true
		}
	}

	if !deleted {
		return fmt.Errorf("记录不存在: %s", id)
	}

	// 更新记录列表
	records.Records = newRecords

	// 重新构建索引
	records.Index = map[string]map[string][]string{
		"by_flow_id": {},
		"by_status":  {},
		"by_date":    {},
	}

	for _, record := range records.Records {
		updateExecutionRecordIndex(records, record)
	}

	// 保存记录
	return SaveExecutionRecords(filePath, records)
}

// ExecuteFlowByIDRequest 定义通过ID执行流程的请求结构
type ExecuteFlowByIDRequest struct {
	AutoFlowID string                 `json:"autoFlowID"` // 流程ID
	Parameters map[string]interface{} `json:"parameters"` // 流程参数
	Process    bool                   `json:"process"`    // 是否返回详细执行日志
}

// ExecuteFlowByIDResponse 定义通过ID执行流程的响应结构
type ExecuteFlowByIDResponse struct {
	Success       bool                     `json:"success"`                  // 是否成功
	InstanceID    string                   `json:"instance_id,omitempty"`    // 流程实例ID
	Logs          []map[string]interface{} `json:"logs"`                     // 执行日志
	FinalOutput   interface{}              `json:"final_output,omitempty"`   // 最终输出
	ExecutedNodes int                      `json:"executed_nodes,omitempty"` // 执行的节点数
	Error         string                   `json:"error,omitempty"`          // 错误信息
}

// 模拟的工作流存储（实际应该使用数据库）
var flowStore = make(map[string]*workflow.FlowDefinition)

// 初始化一些示例工作流
func init() {
	flowStore["flow_1"] = &workflow.FlowDefinition{
		ID:          "flow_1",
		Name:        "测试流程1",
		Description: "这是一个测试流程",
		StartNode:   "start_1",
		Nodes: []workflow.Node{
			{ID: "start_1", Type: "start", Properties: make(map[string]interface{}), Next: "end_1"},
			{ID: "end_1", Type: "end", Properties: make(map[string]interface{}), Next: ""},
		},
	}

	flowStore["flow_2"] = &workflow.FlowDefinition{
		ID:          "flow_2",
		Name:        "测试流程2",
		Description: "这是另一个测试流程",
		StartNode:   "start_2",
		Nodes: []workflow.Node{
			{ID: "start_2", Type: "start", Properties: make(map[string]interface{}), Next: "http_1"},
			{ID: "http_1", Type: "http", Properties: map[string]interface{}{
				"url":     "https://jsonplaceholder.typicode.com/posts/1",
				"method":  "GET",
				"timeout": 30000,
			}, Next: "end_2"},
			{ID: "end_2", Type: "end", Properties: make(map[string]interface{}), Next: ""},
		},
	}
}

// ExecuteFlowHandler 处理流程执行请求
func ExecuteFlowHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许POST请求
	if r.Method != http.MethodPost {
		http.Error(w, "只允许POST请求", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("读取请求体失败: %v", err)
		http.Error(w, "读取请求体失败", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 解析请求参数
	var req FlowExecuteRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("解析请求参数失败: %v", err)
		http.Error(w, "解析请求参数失败", http.StatusBadRequest)
		return
	}

	// 验证请求参数
	if req.Flow == nil {
		http.Error(w, "流程定义不能为空", http.StatusBadRequest)
		return
	}

	// 创建流程解析器
	parser := workflow.NewJSONFlowParser()

	// 验证流程
	if err := parser.Validate(req.Flow); err != nil {
		log.Printf("流程验证失败: %v", err)
		http.Error(w, fmt.Sprintf("流程验证失败: %v", err), http.StatusBadRequest)
		return
	}

	// 创建流程执行引擎
	engine := createFlowEngine()

	// 注册流程
	engine.RegisterFlow(req.Flow)

	// 执行流程
	ctx := context.Background()
	initialContext := &workflow.Context{
		Data: make(map[string]interface{}),
	}

	// 直接调用executeFlowInstance方法获取详细执行结果
	// 注意：这里需要将SimpleFlowEngine转换为具体类型，因为FlowEngine接口没有暴露executeFlowInstance方法
	simpleEngine, ok := engine.(*workflow.SimpleFlowEngine)
	if !ok {
		log.Printf("转换流程引擎类型失败")
		http.Error(w, "转换流程引擎类型失败", http.StatusInternalServerError)
		return
	}

	// 记录执行结果
	startTime := time.Now()
	executionID := fmt.Sprintf("exec_%d", startTime.UnixMilli())

	// 构建错误信息列表
	errors := []ExecutionError{}

	// 创建初始执行记录（状态为running）
	initialRecord := ExecutionRecord{
		ID:            executionID,
		FlowID:        req.Flow.ID,
		FlowName:      req.Flow.Name,
		StartTime:     startTime.Format(time.RFC3339),
		EndTime:       "",
		Status:        "running",
		DurationMs:    0,
		Input:         initialContext.Data,
		Output:        map[string]interface{}{},
		NodesExecuted: 0,
		Errors:        errors,
	}

	// 保存初始执行记录
	dataFilePath := getDataFilePath("execution_records.json")
	if err := AddExecutionRecord(dataFilePath, initialRecord); err != nil {
		log.Printf("保存初始执行记录失败: %v", err)
	}

	// 执行流程并获取详细结果
	result := simpleEngine.ExecuteFlowInstance(ctx, &workflow.FlowInstance{
		ID:     executionID, // 使用执行记录ID作为实例ID
		FlowID: req.Flow.ID,
	}, req.Flow, initialContext)
	endTime := time.Now()
	durationMs := int(endTime.Sub(startTime).Milliseconds())

	// 添加调试日志
	log.Printf("流程执行完成，Success: %v, Logs数量: %d, Error: %s", result.Success, len(result.Logs), result.Error)

	// 构建错误信息列表
	errors = []ExecutionError{}
	if result.Error != "" {
		errors = append(errors, ExecutionError{
			NodeID:  "flow",
			Message: result.Error,
		})
	}

	// 设置执行状态
	status := "success"
	if !result.Success {
		status = "error"
	}

	// 安全地转换 Output 字段
	var outputMap map[string]interface{}
	if result.FinalOutput != nil {
		// 尝试类型断言
		if m, ok := result.FinalOutput.(map[string]interface{}); ok {
			outputMap = m
		} else {
			// 如果类型断言失败，创建一个包含 value 字段的 map
			outputMap = map[string]interface{}{
				"value": result.FinalOutput,
			}
		}
	} else {
		// 如果 FinalOutput 为 nil，使用空 map
		outputMap = map[string]interface{}{}
	}

	// 加载现有记录
	records, err := LoadExecutionRecords(dataFilePath)
	if err != nil {
		log.Printf("加载执行记录失败: %v", err)
	} else {
		// 更新执行记录
		recordUpdated := false
		for i, record := range records.Records {
			if record.ID == executionID {
				records.Records[i] = ExecutionRecord{
					ID:            executionID,
					FlowID:        req.Flow.ID,
					FlowName:      req.Flow.Name,
					StartTime:     startTime.Format(time.RFC3339),
					EndTime:       endTime.Format(time.RFC3339),
					Status:        status,
					DurationMs:    durationMs,
					Input:         initialContext.Data,
					Output:        outputMap,
					NodesExecuted: result.ExecutedNodes,
					Errors:        errors,
					Logs:          convertLogsToMap(result.Logs),
				}
				recordUpdated = true
				break
			}
		}

		// 如果没有找到记录，添加新记录
		if !recordUpdated {
			log.Printf("没有找到初始执行记录，添加新记录")
			newRecord := ExecutionRecord{
				ID:            executionID,
				FlowID:        req.Flow.ID,
				FlowName:      req.Flow.Name,
				StartTime:     startTime.Format(time.RFC3339),
				EndTime:       endTime.Format(time.RFC3339),
				Status:        status,
				DurationMs:    durationMs,
				Input:         initialContext.Data,
				Output:        outputMap,
				NodesExecuted: result.ExecutedNodes,
				Errors:        errors,
				Logs:          convertLogsToMap(result.Logs),
			}
			records.Records = append(records.Records, newRecord)
		}

		// 重新构建索引
		records.Index = map[string]map[string][]string{
			"by_flow_id": {},
			"by_status":  {},
			"by_date":    {},
		}

		for _, record := range records.Records {
			updateExecutionRecordIndex(records, record)
		}

		// 保存更新后的记录
		if err := SaveExecutionRecords(dataFilePath, records); err != nil {
			log.Printf("保存更新后的执行记录失败: %v", err)
		} else {
			log.Printf("执行记录保存成功: ID=%s, FlowID=%s, Status=%s", executionID, req.Flow.ID, status)
		}
	}

	// 构造HTTP响应
	response := FlowExecuteResponse{
		Success:       result.Success,
		InstanceID:    executionID, // 使用执行记录ID作为实例ID
		Logs:          convertLogsToMap(result.Logs),
		FinalOutput:   result.FinalOutput,
		ExecutedNodes: result.ExecutedNodes,
		Error:         result.Error,
	}

	// 添加调试日志
	log.Printf("构造HTTP响应完成，Logs长度: %d", len(response.Logs))

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	if result.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(response)
}

// GetFlowsHandler 处理获取工作流列表请求
func GetFlowsHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}

	// 调用外部API获取工作流列表
	client := httpclient.NewClient()

	// 构建请求配置
	config := &httpclient.RequestConfig{
		URL:         "https://xmzail.com/autoSet/CCAM/autoDataSelectAPI?token=QMYG88888&autoDataKey=1148",
		Method:      httpclient.RequestMethod("GET"),
		ContentType: httpclient.ContentType("application/json"),
		Timeout:     30000,
	}

	// 发送请求
	resp, err := client.Do(config)
	if err != nil {
		log.Printf("调用外部API失败: %v", err)
		http.Error(w, fmt.Sprintf("调用外部API失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 直接返回外部API的响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// 根据Body的实际类型进行处理
	switch body := resp.Body.(type) {
	case string:
		// 如果是字符串，直接写入响应
		w.Write([]byte(body))
	default:
		// 否则，将其编码为JSON
		if err := json.NewEncoder(w).Encode(body); err != nil {
			log.Printf("编码响应失败: %v", err)
			http.Error(w, "编码响应失败", http.StatusInternalServerError)
			return
		}
	}
}

// ExecuteFlowByIDHandler 处理通过ID执行流程的请求
func ExecuteFlowByIDHandler(w http.ResponseWriter, r *http.Request) {
	// 支持GET和POST请求
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "只允许GET或POST请求", http.StatusMethodNotAllowed)
		return
	}

	var req ExecuteFlowByIDRequest

	// 检查是否使用路径参数
	path := r.URL.Path
	if strings.HasPrefix(path, "/api/flow/execute-by-id/") {
		// 从路径中提取 ID
		id := strings.TrimPrefix(path, "/api/flow/execute-by-id/")
		if id != "" {
			req.AutoFlowID = id
		}
	} else if strings.HasPrefix(path, "/workflow/execute/") {
		// 从新路径格式中提取 ID
		id := strings.TrimPrefix(path, "/workflow/execute/")
		if id != "" {
			req.AutoFlowID = id
		}
	}

	// 根据请求方法处理参数
	if r.Method == http.MethodGet {
		// GET请求：从URL查询参数获取
		query := r.URL.Query()

		// 如果路径参数中没有获取到 autoFlowID，则从查询参数获取
		if req.AutoFlowID == "" {
			req.AutoFlowID = query.Get("autoFlowID")
		}

		// 处理process参数
		if processStr := query.Get("process"); processStr == "true" || processStr == "1" {
			req.Process = true
		}

		// 使用GetParamsByQuery方法获取参数，剔除autoFlowID和process字段
		params, err := GetParamsByQuery(query, "autoFlowID", "process")
		if err != nil {
			log.Printf("获取参数失败: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 将获取的参数设置到请求中
		req.Parameters = params
	} else {
		// POST请求：从请求体获取
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("读取请求体失败: %v", err)
			http.Error(w, "读取请求体失败", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// 解析请求参数
		if err := json.Unmarshal(body, &req); err != nil {
			log.Printf("解析请求参数失败: %v", err)
			http.Error(w, "解析请求参数失败", http.StatusBadRequest)
			return
		}
	}

	// 验证请求参数
	if req.AutoFlowID == "" {
		http.Error(w, "autoFlowID不能为空", http.StatusBadRequest)
		return
	}

	// 通过autoFlowID获取流程配置
	flowDef, err := getFlowDefinitionByAutoFlowID(req.AutoFlowID, req.Parameters)
	if err != nil {
		log.Printf("获取流程配置失败: %v", err)
		http.Error(w, fmt.Sprintf("获取流程配置失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 创建流程执行引擎
	engine := createFlowEngine()

	// 注册流程
	engine.RegisterFlow(flowDef)

	// 执行流程 - 设置 5 分钟超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	initialContext := &workflow.Context{
		Data: make(map[string]interface{}),
	}

	// 将请求参数赋值到上下文数据中
	for key, value := range req.Parameters {
		initialContext.Data[key] = value
	}

	// 设置 process 参数，让子工作流继承
	initialContext.Process = req.Process

	// 直接调用executeFlowInstance方法获取详细执行结果
	simpleEngine, ok := engine.(*workflow.SimpleFlowEngine)
	if !ok {
		log.Printf("转换流程引擎类型失败")
		http.Error(w, "转换流程引擎类型失败", http.StatusInternalServerError)
		return
	}

	// 记录执行结果
	startTime := time.Now()
	executionID := fmt.Sprintf("exec_%d", startTime.UnixMilli())

	// 构建错误信息列表
	errors := []ExecutionError{}

	// 创建初始执行记录（状态为running）
	initialRecord := ExecutionRecord{
		ID:            executionID,
		FlowID:        flowDef.ID,
		FlowName:      flowDef.Name,
		StartTime:     startTime.Format(time.RFC3339),
		EndTime:       "",
		Status:        "running",
		DurationMs:    0,
		Input:         req.Parameters,
		Output:        map[string]interface{}{},
		NodesExecuted: 0,
		Errors:        errors,
	}

	// 保存初始执行记录
	dataFilePath := getDataFilePath("execution_records.json")
	if err := AddExecutionRecord(dataFilePath, initialRecord); err != nil {
		log.Printf("保存初始执行记录失败: %v", err)
	}

	// 执行流程并获取详细结果
	result := simpleEngine.ExecuteFlowInstance(ctx, &workflow.FlowInstance{
		ID:     executionID, // 使用执行记录ID作为实例ID
		FlowID: flowDef.ID,
	}, flowDef, initialContext)
	endTime := time.Now()
	durationMs := int(endTime.Sub(startTime).Milliseconds())

	// 构建错误信息列表
	errors = []ExecutionError{}
	if result.Error != "" {
		errors = append(errors, ExecutionError{
			NodeID:  "flow",
			Message: result.Error,
		})
	}

	// 设置执行状态
	status := "success"
	if !result.Success {
		status = "error"
	}

	// 安全地转换 Output 字段
	var outputMap map[string]interface{}
	if result.FinalOutput != nil {
		// 尝试类型断言
		if m, ok := result.FinalOutput.(map[string]interface{}); ok {
			outputMap = m
		} else {
			// 如果类型断言失败，创建一个包含 value 字段的 map
			outputMap = map[string]interface{}{
				"value": result.FinalOutput,
			}
		}
	} else {
		// 如果 FinalOutput 为 nil，使用空 map
		outputMap = map[string]interface{}{}
	}

	// 加载现有记录
	records, err := LoadExecutionRecords(dataFilePath)
	if err != nil {
		log.Printf("加载执行记录失败: %v", err)
	} else {
		// 更新执行记录
		recordUpdated := false
		for i, record := range records.Records {
			if record.ID == executionID {
				records.Records[i] = ExecutionRecord{
					ID:            executionID,
					FlowID:        flowDef.ID,
					FlowName:      flowDef.Name,
					StartTime:     startTime.Format(time.RFC3339),
					EndTime:       endTime.Format(time.RFC3339),
					Status:        status,
					DurationMs:    durationMs,
					Input:         req.Parameters,
					Output:        outputMap,
					NodesExecuted: result.ExecutedNodes,
					Errors:        errors,
					Logs:          convertLogsToMap(result.Logs),
				}
				recordUpdated = true
				break
			}
		}

		// 如果没有找到记录，添加新记录
		if !recordUpdated {
			log.Printf("没有找到初始执行记录，添加新记录")
			newRecord := ExecutionRecord{
				ID:            executionID,
				FlowID:        flowDef.ID,
				FlowName:      flowDef.Name,
				StartTime:     startTime.Format(time.RFC3339),
				EndTime:       endTime.Format(time.RFC3339),
				Status:        status,
				DurationMs:    durationMs,
				Input:         req.Parameters,
				Output:        outputMap,
				NodesExecuted: result.ExecutedNodes,
				Errors:        errors,
				Logs:          convertLogsToMap(result.Logs),
			}
			records.Records = append(records.Records, newRecord)
		}

		// 重新构建索引
		records.Index = map[string]map[string][]string{
			"by_flow_id": {},
			"by_status":  {},
			"by_date":    {},
		}

		for _, record := range records.Records {
			updateExecutionRecordIndex(records, record)
		}

		// 保存更新后的记录
		if err := SaveExecutionRecords(dataFilePath, records); err != nil {
			log.Printf("保存更新后的执行记录失败: %v", err)
		} else {
			log.Printf("执行记录保存成功: ID=%s, FlowID=%s, Status=%s", executionID, flowDef.ID, status)
		}
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")

	// 构造响应（无论成功还是失败都包含 logs）
	response := ExecuteFlowByIDResponse{
		Success:       result.Success,
		InstanceID:    executionID, // 使用执行记录ID作为实例ID
		Logs:          convertLogsToMap(result.Logs),
		FinalOutput:   result.FinalOutput,
		ExecutedNodes: result.ExecutedNodes,
		Error:         result.Error,
	}

	// 根据成功状态设置状态码
	if !result.Success {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// 只有当process=true时才返回详细日志和完整结构
	if req.Process {
		json.NewEncoder(w).Encode(response)
	} else {
		json.NewEncoder(w).Encode(result.FinalOutput)
	}
}

// convertLogsToMap 将ExecutionLog数组转换为map数组
func convertLogsToMap(logs []workflow.ExecutionLog) []map[string]interface{} {
	result := make([]map[string]interface{}, len(logs))
	for i, log := range logs {
		result[i] = map[string]interface{}{
			"timestamp": log.Timestamp,
			"level":     log.Level,
			"log_type":  log.LogType,
			"nodeName":  log.NodeName,
			"content":   log.Content,
		}
	}
	return result
}

// getFlowDefinitionByAutoFlowID 通过uuid_workflow获取流程配置
func getFlowDefinitionByAutoFlowID(uuidWorkflow string, parameters map[string]interface{}) (*workflow.FlowDefinition, error) {
	workflowData, err := GetAutoSet(uuidWorkflow)
	if err != nil {
		return nil, err
	}

	var flowDef *workflow.FlowDefinition

	if autosetData, ok := workflowData["autoset"].(map[string]interface{}); ok {
		auto := autosetData
		name := ""
		if nameValue, ok := auto["name"]; ok && nameValue != nil {
			name = fmt.Sprintf("%v", nameValue)
		}

		description := ""
		if descValue, ok := auto["description"]; ok && descValue != nil {
			description = fmt.Sprintf("%v", descValue)
		}

		startNode := ""
		if startNodeValue, ok := auto["start_node"]; ok && startNodeValue != nil {
			startNode = fmt.Sprintf("%v", startNodeValue)
		}

		flowDef = &workflow.FlowDefinition{
			ID:          uuidWorkflow,
			Name:        name,
			Description: description,
			StartNode:   startNode,
			CreatedAt:   "",
			UpdatedAt:   "",
		}

		if nodes, ok := auto["nodes"].([]interface{}); ok {
			for _, node := range nodes {
				if nodeMap, ok := node.(map[string]interface{}); ok {
					properties := make(map[string]interface{})
					if props, ok := nodeMap["properties"].(map[string]interface{}); ok {
						properties = props
					}

					for key, value := range parameters {
						properties[key] = value
					}

					var position *workflow.Position
					if pos, ok := nodeMap["position"].(map[string]interface{}); ok {
						x, _ := pos["x"].(float64)
						y, _ := pos["y"].(float64)
						position = &workflow.Position{X: x, Y: y}
					}

					var sources []string
					if sourcesValue, exists := nodeMap["sources"]; exists {
						switch v := sourcesValue.(type) {
						case []string:
							sources = v
						case []interface{}:
							for _, item := range v {
								if str, ok := item.(string); ok {
									sources = append(sources, str)
								}
							}
						case string:
							var sourcesArray []string
							if err := json.Unmarshal([]byte(v), &sourcesArray); err == nil {
								sources = sourcesArray
							}
						}
					}

					flowNode := workflow.Node{
						ID:         fmt.Sprintf("%v", nodeMap["id"]),
						Type:       fmt.Sprintf("%v", nodeMap["type"]),
						Properties: properties,
						Position:   position,
						Sources:    sources,
					}

					flowDef.Nodes = append(flowDef.Nodes, flowNode)
				}
			}
		}
	}

	if flowDef == nil {
		return nil, fmt.Errorf("未找到指定的流程")
	}

	return flowDef, nil
}

// HTTPFlowExecutor HTTP工作流执行器实现
type HTTPFlowExecutor struct{}

// ExecuteFlowByAutoFlowID 根据autoFlowID执行工作流
func (e *HTTPFlowExecutor) ExecuteFlowByAutoFlowID(autoFlowID string, parameters map[string]interface{}, process bool) *workflow.FlowExecutionResult {
	// 获取工作流定义
	flowDef, err := getFlowDefinitionByAutoFlowID(autoFlowID, parameters)
	if err != nil {
		log.Printf("获取流程配置失败: %v", err)
		return &workflow.FlowExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("获取流程配置失败: %v", err),
		}
	}

	// 创建初始上下文
	initialContext := &workflow.Context{
		Data: make(map[string]interface{}),
	}

	// 将参数复制到上下文数据中
	for key, value := range parameters {
		initialContext.Data[key] = value
	}

	// 调用方法2执行工作流
	return workflow.ExecuteFlowWithDefinition(flowDef, initialContext, process)
}

// createFlowEngine 创建标准的公用的流程执行引擎
func createFlowEngine() workflow.FlowEngine {
	engine := workflow.NewSimpleFlowEngine()

	engine.RegisterNodeHandler("start", workflow.NewStartNodeHandler())
	engine.RegisterNodeHandler("end", workflow.NewEndNodeHandler())
	engine.RegisterNodeHandler("http", workflow.NewHTTPNodeHandler())
	engine.RegisterNodeHandler("condition", workflow.NewConditionNodeHandler())
	engine.RegisterNodeHandler("loop", workflow.NewLoopNodeHandler())
	engine.RegisterNodeHandler("log", workflow.NewLogNodeHandler())
	engine.RegisterNodeHandler("assign", workflow.NewAssignNodeHandler())
	engine.RegisterNodeHandler("coze_workflow", workflow.NewCozeWorkflowNodeHandler())
	engine.RegisterNodeHandler("string_to_json", workflow.NewStringToJsonNodeHandler())
	engine.RegisterNodeHandler("json_to_string", workflow.NewJsonToStringNodeHandler())
	engine.RegisterNodeHandler("data_fetch", workflow.NewDataFetchNodeHandler())
	engine.RegisterNodeHandler("data_insert", workflow.NewDataInsertNodeHandler())
	engine.RegisterNodeHandler("flow_call", workflow.NewFlowCallNodeHandler())
	engine.RegisterNodeHandler("dingtalk", workflow.NewDingTalkNodeHandler())
	engine.RegisterNodeHandler("plugin", workflow.NewPluginNodeHandler())
	engine.RegisterNodeHandler("sms", workflow.NewSMSNodeHandler())

	// 注册worker节点处理器
	engine.RegisterNodeHandler("worker", workflow.NewWorkerNodeHandler())

	return engine
}

// HomeHandler 处理首页请求
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// 返回简单的HTML页面
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HTTP代理服务</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            line-height: 1.6;
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #4CAF50;
            padding-bottom: 10px;
        }
        h2 {
            color: #4CAF50;
        }
        pre {
            background-color: #f5f5f5;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
        }
        .endpoint {
            background-color: #e8f5e8;
            padding: 10px;
            border-radius: 5px;
            margin: 10px 0;
        }
        .method {
            font-weight: bold;
            color: #4CAF50;
        }
        .example {
            background-color: #f0f8ff;
            padding: 15px;
            border-radius: 5px;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <h1>HTTP代理服务</h1>
    
    <p>这是一个HTTP代理服务，可以通过API调用实现各种HTTP请求的转发和流程执行。</p>
    
    <h2>API端点</h2>
    <div class="endpoint">
        <span class="method">POST</span> /api/proxy - HTTP代理请求
    </div>
    <div class="endpoint">
        <span class="method">POST</span> /api/flow/execute - 执行流程
    </div>
    
    <h2>请求参数</h2>
    <h3>1. HTTP代理请求</h3>
    <pre>
{
    "url": "请求URL",
    "method": "请求方法(GET, POST, PUT, DELETE, PATCH)",
    "content_type": "Content-Type(可选)",
    "headers": {"请求头1": "值1", "请求头2": "值2"}(可选),
    "body": 请求体数据(可选),
    "timeout": 超时时间(毫秒，可选)
}
    </pre>
    
    <h3>2. 流程执行请求</h3>
    <pre>
{
    "flow": {
        "id": "flow_id",
        "name": "流程名称",
        "start_node": "start_node_id",
        "nodes": [
            {
                "id": "start_node_id",
                "type": "start",
                "name": "开始节点",
                "properties": {},
                "next": "next_node_id"
            }
            // 其他节点...
        ]
    }
}
    </pre>
    
    <h2>响应格式</h2>
    <pre>
{
    "success": true/false,
    "result": { /* 响应结果 */ },
    "error": "错误信息(仅当success为false时存在)"
}
    </pre>
    
    <h2>示例请求</h2>
    <div class="example">
            <h3>使用curl发送GET请求</h3>
            <pre>curl -X POST http://localhost:2427/api/proxy \
  -H "Content-Type: application/json" \
  -d '{"url": "https://jsonplaceholder.typicode.com/posts/1", "method": "GET"}'</pre>
        </div>
        
        <div class="example">
            <h3>使用curl发送POST请求</h3>
            <pre>curl -X POST http://localhost:2427/api/proxy \
  -H "Content-Type: application/json" \
  -d '{"url": "https://jsonplaceholder.typicode.com/posts", "method": "POST", "content_type": "application/json", "body": {"title": "测试标题", "body": "测试内容", "userId": 1}}'</pre>
        </div>
</body>
</html>
    `

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// PluginListHandler 处理获取插件列表请求
func PluginListHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}

	// 获取插件管理器实例
	pluginManager := plugin.NewDefaultPluginManager()

	// 注册所有插件
	plugin.RegisterAllPlugins(pluginManager)

	// 获取所有插件定义
	plugins := pluginManager.ListPlugins()

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plugins)
}

// PluginExecuteHandler 处理插件执行请求
func PluginExecuteHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(PluginExecuteResponse{
			Success: false,
			Error:   "读取请求体失败: " + err.Error(),
		})
		return
	}

	// 解析请求
	var req PluginExecuteRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(PluginExecuteResponse{
			Success: false,
			Error:   "解析请求失败: " + err.Error(),
		})
		return
	}

	// 验证插件ID
	if req.PluginID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(PluginExecuteResponse{
			Success: false,
			Error:   "插件ID不能为空",
		})
		return
	}

	// 设置默认超时
	timeout := 30000
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	// 创建插件配置
	pluginConfig := &plugin.PluginConfig{
		PluginID:   req.PluginID,
		Parameters: req.Parameters,
		Timeout:    timeout,
	}

	// 获取插件管理器实例
	pluginManager := plugin.NewDefaultPluginManager()

	// 注册所有插件（包括阿里云短信服务插件）
	plugin.RegisterAllPlugins(pluginManager)

	// 获取插件执行器
	executor, err := pluginManager.GetPlugin(req.PluginID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(PluginExecuteResponse{
			Success: false,
			Error:   "获取插件失败: " + err.Error(),
		})
		return
	}

	// 执行插件
	result, err := executor.Execute(r.Context(), pluginConfig)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(PluginExecuteResponse{
			Success: false,
			Error:   "执行插件失败: " + err.Error(),
		})
		return
	}

	if !result.Success {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(PluginExecuteResponse{
			Success: false,
			Error:   "插件执行失败: " + result.Error,
		})
		return
	}

	// 返回成功响应
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PluginExecuteResponse{
		Success: true,
		Outputs: result.Outputs,
	})
}

// DocsPublicListHandler 处理获取公开文档列表请求
func DocsPublicListHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}

	// 读取docs-public-list.json文件
	filePath := getProjectFilePath("docs-public/docs-public-list.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取文档列表文件失败: %v", err)
		http.Error(w, "读取文档列表失败", http.StatusInternalServerError)
		return
	}

	// 解析文档列表
	var docs []map[string]interface{}
	if err := json.Unmarshal(data, &docs); err != nil {
		log.Printf("解析文档列表失败: %v", err)
		http.Error(w, "解析文档列表失败", http.StatusInternalServerError)
		return
	}

	// 构建响应结构
	response := map[string]interface{}{
		"list":   docs,
		"manual": "如何获取文档详情内容",
	}

	// 编码响应
	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("编码响应失败: %v", err)
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

// DocsPublicGetHandler 处理获取公开文档内容请求
func DocsPublicGetHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}

	// 获取fileName参数
	fileName := r.URL.Query().Get("fileName")
	if fileName == "" {
		http.Error(w, "fileName参数不能为空", http.StatusBadRequest)
		return
	}

	// 构建文件路径
	filePath := "../../docs-public/" + fileName

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取文档文件失败: %v", err)
		http.Error(w, "读取文档失败", http.StatusInternalServerError)
		return
	}

	// 读取docs-public-list.json文件，获取文档的title和desc
	listFilePath := "../../docs-public/docs-public-list.json"
	listData, err := os.ReadFile(listFilePath)
	if err != nil {
		log.Printf("读取文档列表文件失败: %v", err)
		// 即使读取列表失败，也继续返回文档内容
	}

	// 解析文档列表，获取当前文档的title和desc
	title := ""
	desc := ""
	var docs []map[string]interface{}
	if err := json.Unmarshal(listData, &docs); err == nil {
		for _, doc := range docs {
			if docFileName, ok := doc["fileName"].(string); ok && docFileName == fileName {
				if docTitle, ok := doc["title"].(string); ok {
					title = docTitle
				}
				if docDesc, ok := doc["desc"].(string); ok {
					desc = docDesc
				}
				break
			}
		}
	}

	// 构建响应结构
	response := map[string]interface{}{
		"content": string(data),
		"title":   title,
		"desc":    desc,
	}

	// 编码响应
	responseData, err := json.Marshal(response)
	if err != nil {
		log.Printf("编码响应失败: %v", err)
		http.Error(w, "编码响应失败", http.StatusInternalServerError)
		return
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

// CustomNodesConfigHandler 处理获取自定义节点配置请求
func CustomNodesConfigHandler(w http.ResponseWriter, r *http.Request) {
	// 只允许GET请求
	if r.Method != http.MethodGet {
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}

	// 构建文件路径
	filePath := getProjectFilePath("frontend/src/assets/custom-nodes/custom-node-config.json")

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取自定义节点配置文件失败: %v", err)
		http.Error(w, "读取自定义节点配置失败", http.StatusInternalServerError)
		return
	}

	// 解析JSON配置
	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Printf("解析自定义节点配置失败: %v", err)
		http.Error(w, "解析自定义节点配置失败", http.StatusInternalServerError)
		return
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// GetParamsByQuery 从URL查询参数中获取参数，排除指定的字段
func GetParamsByQuery(query map[string][]string, excludeFields ...string) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// 创建排除字段的映射，方便快速查找
	excludeMap := make(map[string]bool)
	for _, field := range excludeFields {
		excludeMap[field] = true
	}

	// 遍历查询参数
	for key, values := range query {
		// 跳过排除的字段
		if excludeMap[key] {
			continue
		}

		// 如果参数只有一个值，直接使用该值
		if len(values) == 1 {
			params[key] = values[0]
		} else {
			// 如果参数有多个值，使用值的切片
			params[key] = values
		}
	}

	return params, nil
}

// ExecutionRecordHandler 处理工作流执行记录请求
func ExecutionRecordHandler(w http.ResponseWriter, r *http.Request) {
	// 构建数据文件路径
	dataFilePath := getDataFilePath("execution_records.json")

	// 辅助函数：移除 logs 字段
	removeLogsFromRecords := func(records []ExecutionRecord) []map[string]interface{} {
		result := make([]map[string]interface{}, len(records))
		for i, record := range records {
			// 序列化再反序列化，转换为 map
			recordBytes, _ := json.Marshal(record)
			var recordMap map[string]interface{}
			json.Unmarshal(recordBytes, &recordMap)
			// 移除 logs 字段
			delete(recordMap, "logs")
			result[i] = recordMap
		}
		return result
	}

	switch r.Method {
	case http.MethodGet:
		// 处理查询请求
		query := r.URL.Query()

		// 检查查询参数
		if flowID := query.Get("flow_id"); flowID != "" {
			// 按工作流ID查询
			records, err := GetExecutionRecordsByFlowID(dataFilePath, flowID)
			if err != nil {
				log.Printf("查询工作流执行记录失败: %v", err)
				http.Error(w, "查询工作流执行记录失败", http.StatusInternalServerError)
				return
			}

			// 返回JSON响应（不带 logs）
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(removeLogsFromRecords(records))
			return
		}

		if status := query.Get("status"); status != "" {
			// 按状态查询
			records, err := GetExecutionRecordsByStatus(dataFilePath, status)
			if err != nil {
				log.Printf("查询执行记录失败: %v", err)
				http.Error(w, "查询执行记录失败", http.StatusInternalServerError)
				return
			}

			// 返回JSON响应（不带 logs）
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(removeLogsFromRecords(records))
			return
		}

		if date := query.Get("date"); date != "" {
			// 按日期查询
			records, err := GetExecutionRecordsByDate(dataFilePath, date)
			if err != nil {
				log.Printf("查询执行记录失败: %v", err)
				http.Error(w, "查询执行记录失败", http.StatusInternalServerError)
				return
			}

			// 返回JSON响应（不带 logs）
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(removeLogsFromRecords(records))
			return
		}

		if id := query.Get("id"); id != "" {
			// 按ID查询（保留 logs）
			record, err := GetExecutionRecordByID(dataFilePath, id)
			if err != nil {
				log.Printf("查询执行记录失败: %v", err)
				http.Error(w, "查询执行记录失败", http.StatusInternalServerError)
				return
			}

			// 返回JSON响应（保留 logs）
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(record)
			return
		}

		// 获取所有记录
		records, err := GetAllExecutionRecords(dataFilePath)
		if err != nil {
			log.Printf("查询所有执行记录失败: %v", err)
			http.Error(w, "查询所有执行记录失败", http.StatusInternalServerError)
			return
		}

		// 返回JSON响应（不带 logs）
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(removeLogsFromRecords(records))
		return

	case http.MethodPost:
		// 处理添加记录请求
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("读取请求体失败: %v", err)
			http.Error(w, "读取请求体失败", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// 解析请求体
		var record ExecutionRecord
		if err := json.Unmarshal(body, &record); err != nil {
			log.Printf("解析请求体失败: %v", err)
			http.Error(w, "解析请求体失败", http.StatusBadRequest)
			return
		}

		// 添加记录
		if err := AddExecutionRecord(dataFilePath, record); err != nil {
			log.Printf("添加执行记录失败: %v", err)
			http.Error(w, "添加执行记录失败", http.StatusInternalServerError)
			return
		}

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "执行记录添加成功",
		})
		return

	case http.MethodDelete:
		// 处理删除记录请求
		query := r.URL.Query()
		id := query.Get("id")

		if id == "" {
			http.Error(w, "ID参数不能为空", http.StatusBadRequest)
			return
		}

		// 删除记录
		if err := DeleteExecutionRecord(dataFilePath, id); err != nil {
			log.Printf("删除执行记录失败: %v", err)
			http.Error(w, "删除执行记录失败", http.StatusInternalServerError)
			return
		}

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "执行记录删除成功",
		})
		return

	default:
		http.Error(w, "只允许GET、POST和DELETE请求", http.StatusMethodNotAllowed)
		return
	}
}

func GetAutosetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}

	path := r.URL.Path
	uuidWorkflow := ""
	if strings.HasPrefix(path, "/workflow/api/autoset/") {
		uuidWorkflow = strings.TrimPrefix(path, "/workflow/api/autoset/")
	}

	if uuidWorkflow == "" {
		uuidWorkflow = r.URL.Query().Get("uuid_workflow")
	}

	if uuidWorkflow == "" {
		http.Error(w, "uuid_workflow参数不能为空", http.StatusBadRequest)
		return
	}

	workflowData, err := GetAutoSet(uuidWorkflow)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    workflowData,
	})
}

func GetWorkflowListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只允许GET请求", http.StatusMethodNotAllowed)
		return
	}

	workflowList, err := GetWorkflowList()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    workflowList,
	})
}
