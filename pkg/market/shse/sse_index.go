package shse

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// 上海证券交易所指数列表接口地址
	urlSSEIndex = "https://query.sse.com.cn/commonSoaQuery.do"
	// Referer地址，模拟浏览器请求，避免被接口拒绝
	referer = "http://www.sse.com.cn/"
)

// SSEIndexItem 上海证券交易所指数列表项结构体
// 用于解析接口返回的指数数据
type SSEIndexItem struct {
	IndexCode    string `json:"INDEX_CODE"`    // 指数代码
	IndexName    string `json:"INDEX_NAME"`    // 指数名称
	IndexEnglish string `json:"INDEX_ENGLISH"` // 指数英文名称
	BasePoint    string `json:"BASE_POINT"`    // 基点
	BaseDate     string `json:"BASE_DATE"`     // 基期
	CalcMethod   string `json:"CALC_METHOD"`   // 计算方法
	PubDate      string `json:"PUB_DATE"`      // 发布日期
	LastModify   string `json:"LAST_MODIFY"`   // 最近修改日期
	IndexIntro   string `json:"INDEX_INTRO"`   // 指数简介
}

// SSEIndexResponse 上海证券交易所指数接口返回结构体
type SSEIndexResponse struct {
	Result  []SSEIndexItem `json:"result"`  // 指数数据列表
	Success bool           `json:"success"` // 请求是否成功
	Error   struct {
		ErrorCode string `json:"errorCode"` // 错误码
		ErrorMsg  string `json:"errorMsg"`  // 错误信息
	} `json:"error"` // 错误信息
}

// IndexList 获取上海证券交易所指数列表
// 返回指数列表、请求耗时（毫秒）和错误信息
func IndexList() ([]SSEIndexItem, int64, error) {
	// 初始化随机数生成器，用于生成jsonp回调函数名的随机数
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 1. 构建请求参数
	now := time.Now()
	timestamp := now.UnixMilli()                         // 当前时间戳（毫秒）
	cbNum := int(math.Floor(rand.Float64() * 100000001)) // 0-100000000的随机数

	// 构建查询参数
	params := url.Values{}
	params.Set("jsonCallBack", fmt.Sprintf("jsonpCallback%d", cbNum)) // jsonp回调函数名
	params.Set("isPagination", "false")                               // 是否分页，false表示获取全部数据
	params.Set("sqlId", "DB_SZZSLB_ZSLB")                             // 接口SQL ID，固定值
	params.Set("_", fmt.Sprintf("%d", timestamp))                     // 时间戳，用于防止缓存

	// 2. 构建完整请求URL
	fullURL := fmt.Sprintf("%s?%s", urlSSEIndex, params.Encode())

	// 3. 创建HTTP请求并设置请求头
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头，模拟浏览器行为，避免接口拦截
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Connection", "keep-alive")

	// 4. 发送请求并记录耗时
	startTime := time.Now()
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置请求超时时间，避免无限等待
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("发送请求失败: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body) // 确保响应体被关闭，避免资源泄漏

	// 计算请求耗时（毫秒）
	costTime := time.Since(startTime).Milliseconds()

	// 5. 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, costTime, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	// 6. 读取响应内容
	var respBody strings.Builder
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			respBody.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	bodyStr := respBody.String()

	// 7. 处理JSONP格式响应（去掉回调函数包裹）
	// JSONP格式：jsonpCallback12345({"success":true,"result":[]})
	cbPrefix := fmt.Sprintf("jsonpCallback%d(", cbNum)
	cbSuffix := ")"

	// 检查响应是否为合法的JSONP格式
	if !strings.HasPrefix(bodyStr, cbPrefix) || !strings.HasSuffix(bodyStr, cbSuffix) {
		return nil, costTime, fmt.Errorf("响应格式不是预期的JSONP格式，内容: %s", bodyStr)
	}

	// 提取JSON部分（去掉前后的回调函数包裹）
	jsonStr := bodyStr[len(cbPrefix) : len(bodyStr)-len(cbSuffix)]

	// 8. 解析JSON数据
	var indexResp SSEIndexResponse
	err = json.Unmarshal([]byte(jsonStr), &indexResp)
	if err != nil {
		return nil, costTime, fmt.Errorf("解析JSON数据失败，JSON内容: %s, 错误: %w", jsonStr, err)
	}

	// 9. 检查请求是否成功
	if !indexResp.Success {
		return nil, costTime, fmt.Errorf("接口请求失败，错误码: %s, 错误信息: %s",
			indexResp.Error.ErrorCode, indexResp.Error.ErrorMsg)
	}

	// 10. 返回指数列表、耗时和无错误
	return indexResp.Result, costTime, nil
}

// PrintIndexList 打印上海证券交易所指数列表（辅助函数）
func PrintIndexList() {
	indexes, costTime, err := IndexList()
	if err != nil {
		fmt.Printf("获取指数列表失败: %v\n", err)
		return
	}

	fmt.Printf("请求耗时: %d 毫秒\n", costTime)
	fmt.Printf("共获取到 %d 个指数\n", len(indexes))
	fmt.Println("=" + strings.Repeat("-", 50) + "=")

	// 打印表头
	fmt.Printf("%-12s %-20s %-10s %-12s\n",
		"指数代码", "指数名称", "基点", "基期")
	fmt.Println("-" + strings.Repeat("-", 50) + "-")

	// 打印每个指数的信息
	for _, index := range indexes {
		fmt.Printf("%-12s %-20s %-10s %-12s\n",
			index.IndexCode,
			index.IndexName,
			index.BasePoint,
			index.BaseDate)
	}

	fmt.Println("=" + strings.Repeat("-", 50) + "=")
}
