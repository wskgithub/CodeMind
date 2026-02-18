# 系统监控仪表盘

系统监控仪表盘为管理员提供实时的系统资源、请求性能和 LLM 节点状态监控。

## 功能特性

### 1. 系统资源监控
- **CPU 使用率**：实时 CPU 利用率、核心数、型号、系统负载
- **内存使用率**：总内存、已用内存、空闲内存、使用率百分比
- **磁盘使用率**：各分区挂载点、设备、使用率、容量信息
- **网络 IO**：网卡收发数据统计

### 2. 请求性能监控
- **QPS (Queries Per Second)**：每秒请求数
- **响应时间**：平均响应时间、P95、P99 延迟
- **错误率**：HTTP 错误率统计
- **状态码分布**：各状态码请求数量统计

### 3. LLM 节点监控
- **节点状态**：在线、离线、忙碌、错误、空闲
- **GPU 信息**：利用率、显存使用、温度
- **资源使用**：CPU 使用率、内存使用率
- **请求统计**：每分钟请求数、活跃请求数、平均响应时间
- **模型信息**：已加载模型列表和数量

## 技术实现

### 后端

#### 数据模型
- `SystemMetric`：系统指标数据（CPU/内存/磁盘/网络/负载）
- `LLMNodeMetric`：LLM 节点上报数据

#### API 接口
| 接口 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/v1/monitor/dashboard` | GET | 获取仪表盘汇总数据 | 超级管理员 |
| `/api/v1/monitor/system` | GET | 获取系统资源指标 | 超级管理员 |
| `/api/v1/monitor/requests` | GET | 获取请求性能指标 | 超级管理员 |
| `/api/v1/monitor/llm-nodes` | GET | 获取 LLM 节点列表 | 超级管理员 |
| `/api/v1/monitor/health` | GET | 健康检查 | 超级管理员 |

#### 系统指标收集
使用 `gopsutil` 库每 30 秒自动收集一次系统资源数据，保留最近 7 天的数据。

### 前端

#### 页面路由
- 路径：`/admin/monitor`
- 权限：仅超级管理员可访问

#### 功能特性
- 自动刷新（支持 5s/10s/30s 间隔）
- 实时数据展示
- 可视化进度条和仪表盘
- 节点状态表格

## LLM 节点数据上报

LLM 节点需要通过 API 定期上报状态数据：

### 上报接口
```
POST /api/v1/monitor/nodes/report
Content-Type: application/json

{
  "node_id": "gpu-node-01",
  "node_name": "GPU 服务器 01",
  "status": "online",
  "gpu_info": [
    {
      "index": 0,
      "name": "NVIDIA A100",
      "total_memory_gb": 80,
      "used_memory_gb": 45,
      "temperature": 65,
      "utilization": 85
    }
  ],
  "gpu_utilization": 85,
  "cpu_cores": 64,
  "cpu_usage_percent": 45,
  "memory_total_gb": 512,
  "memory_used_gb": 256,
  "requests_per_min": 120,
  "avg_response_time_ms": 150,
  "active_requests": 15,
  "queued_requests": 2,
  "loaded_models": [
    {"model_id": "gpt-4", "model_name": "GPT-4", "loaded_at": "2026-02-19T00:00:00Z", "request_count": 1000}
  ],
  "version": "v2.1.0",
  "labels": {"region": "beijing", "zone": "zone-a"},
  "timestamp": 1708272000
}
```

### 上报示例脚本

```python
import requests
import time

# LLM 节点上报示例
report_data = {
    "node_id": "llm-node-01",
    "node_name": "LLM 服务节点 01",
    "status": "online",
    "gpu_utilization": 75.5,
    "cpu_cores": 32,
    "cpu_usage_percent": 45.2,
    "memory_total_gb": 256,
    "memory_used_gb": 128,
    "requests_per_min": 85,
    "avg_response_time_ms": 120,
    "active_requests": 12,
    "queued_requests": 0,
    "version": "v1.0.0",
    "timestamp": int(time.time())
}

response = requests.post(
    "http://your-codemind-server/api/v1/monitor/nodes/report",
    json=report_data
)
print(response.json())
```

## 部署说明

### 1. 数据库迁移
系统会自动创建以下表：
- `system_metrics`：系统指标数据
- `llm_node_metrics`：LLM 节点指标数据

### 2. 依赖安装
```bash
cd backend
go mod tidy  # 自动安装 gopsutil 依赖
```

### 3. 前端依赖
```bash
cd frontend
npm install recharts  # 图表库（可选）
```

## 注意事项

1. **权限控制**：监控仪表盘仅超级管理员可见，已在路由和菜单中配置
2. **数据保留**：系统指标保留 7 天，LLM 节点指标保留 48 小时
3. **性能影响**：系统指标收集每 30 秒执行一次，对服务器性能影响极小
4. **Linux 系统**：gopsutil 在 Linux 系统上支持最完整

## 扩展建议

1. **告警功能**：可根据 CPU/内存/磁盘使用率设置告警阈值
2. **历史趋势**：可添加历史数据趋势图表
3. **多服务器**：支持多服务器监控（通过 hostname 区分）
4. **自定义指标**：可扩展支持自定义业务指标上报
