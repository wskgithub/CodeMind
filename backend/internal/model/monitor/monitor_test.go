package monitor

import (
	"testing"
	"time"
)

// TestSystemMetric_TableName 测试 SystemMetric 表名
func TestSystemMetric_TableName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "should return system_metrics",
			expected: "system_metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := SystemMetric{}
			got := sm.TableName()
			if got != tt.expected {
				t.Errorf("TableName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestSystemMetric_Constants 测试 MetricType 常量
func TestSystemMetric_Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "MetricTypeCPU should be cpu",
			constant: MetricTypeCPU,
			expected: "cpu",
		},
		{
			name:     "MetricTypeMemory should be memory",
			constant: MetricTypeMemory,
			expected: "memory",
		},
		{
			name:     "MetricTypeDisk should be disk",
			constant: MetricTypeDisk,
			expected: "disk",
		},
		{
			name:     "MetricTypeNetwork should be network",
			constant: MetricTypeNetwork,
			expected: "network",
		},
		{
			name:     "MetricTypeLoad should be load",
			constant: MetricTypeLoad,
			expected: "load",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("constant = %v, want %v", tt.constant, tt.expected)
			}
		})
	}
}

// TestSystemMetric_Struct 测试 SystemMetric 结构体创建和字段赋值
func TestSystemMetric_Struct(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		data SystemMetric
		want SystemMetric
	}{
		{
			name: "should create SystemMetric with all fields",
			data: SystemMetric{
				ID:         1,
				HostName:   "server-01",
				MetricType: MetricTypeCPU,
				MetricName: "usage_percent",
				Value:      75.5,
				Labels:     `{"core":"0"}`,
				CreatedAt:  now,
			},
			want: SystemMetric{
				ID:         1,
				HostName:   "server-01",
				MetricType: "cpu",
				MetricName: "usage_percent",
				Value:      75.5,
				Labels:     `{"core":"0"}`,
				CreatedAt:  now,
			},
		},
		{
			name: "should create SystemMetric with memory metric",
			data: SystemMetric{
				ID:         2,
				HostName:   "server-02",
				MetricType: MetricTypeMemory,
				MetricName: "used_gb",
				Value:      16.5,
				Labels:     "",
				CreatedAt:  now,
			},
			want: SystemMetric{
				ID:         2,
				HostName:   "server-02",
				MetricType: "memory",
				MetricName: "used_gb",
				Value:      16.5,
				Labels:     "",
				CreatedAt:  now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", tt.data.ID, tt.want.ID)
			}
			if tt.data.HostName != tt.want.HostName {
				t.Errorf("HostName = %v, want %v", tt.data.HostName, tt.want.HostName)
			}
			if tt.data.MetricType != tt.want.MetricType {
				t.Errorf("MetricType = %v, want %v", tt.data.MetricType, tt.want.MetricType)
			}
			if tt.data.MetricName != tt.want.MetricName {
				t.Errorf("MetricName = %v, want %v", tt.data.MetricName, tt.want.MetricName)
			}
			if tt.data.Value != tt.want.Value {
				t.Errorf("Value = %v, want %v", tt.data.Value, tt.want.Value)
			}
			if tt.data.Labels != tt.want.Labels {
				t.Errorf("Labels = %v, want %v", tt.data.Labels, tt.want.Labels)
			}
		})
	}
}

// TestSystemMetricsSummary_Struct 测试 SystemMetricsSummary 结构体
func TestSystemMetricsSummary_Struct(t *testing.T) {
	now := time.Now()

	cpuMetrics := &CPUMetrics{
		UsagePercent: 45.5,
		CoreCount:    8,
		ModelName:    "Intel i7",
	}

	memoryMetrics := &MemoryMetrics{
		TotalGB:      32.0,
		UsedGB:       16.0,
		FreeGB:       16.0,
		UsagePercent: 50.0,
	}

	diskMetrics := []DiskMetrics{
		{
			MountPoint:   "/",
			Device:       "/dev/sda1",
			TotalGB:      500.0,
			UsedGB:       250.0,
			FreeGB:       250.0,
			UsagePercent: 50.0,
		},
	}

	networkMetrics := &NetworkMetrics{
		InterfaceName: "eth0",
		BytesSentMB:   1024.0,
		BytesRecvMB:   2048.0,
		PacketsSent:   10000,
		PacketsRecv:   20000,
	}

	loadMetrics := &LoadMetrics{
		Load1:  1.5,
		Load5:  1.2,
		Load15: 1.0,
	}

	tests := []struct {
		name string
		data SystemMetricsSummary
	}{
		{
			name: "should create SystemMetricsSummary with all fields",
			data: SystemMetricsSummary{
				CPUUsage:     cpuMetrics,
				MemoryUsage:  memoryMetrics,
				DiskUsage:    diskMetrics,
				NetworkIO:    networkMetrics,
				LoadAverage:  loadMetrics,
				RecordedAt:   now,
			},
		},
		{
			name: "should create SystemMetricsSummary with nil fields",
			data: SystemMetricsSummary{
				CPUUsage:     nil,
				MemoryUsage:  nil,
				DiskUsage:    nil,
				NetworkIO:    nil,
				LoadAverage:  nil,
				RecordedAt:   now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.CPUUsage != cpuMetrics && tt.name == "should create SystemMetricsSummary with all fields" {
				t.Errorf("CPUUsage mismatch")
			}
			if tt.data.MemoryUsage != memoryMetrics && tt.name == "should create SystemMetricsSummary with all fields" {
				t.Errorf("MemoryUsage mismatch")
			}
			if tt.data.RecordedAt != now {
				t.Errorf("RecordedAt mismatch")
			}
		})
	}
}

// TestCPUMetrics_Struct 测试 CPUMetrics 结构体
func TestCPUMetrics_Struct(t *testing.T) {
	tests := []struct {
		name     string
		data     CPUMetrics
		expected CPUMetrics
	}{
		{
			name: "should create CPUMetrics with valid values",
			data: CPUMetrics{
				UsagePercent: 75.5,
				CoreCount:    16,
				ModelName:    "AMD Ryzen 9",
			},
			expected: CPUMetrics{
				UsagePercent: 75.5,
				CoreCount:    16,
				ModelName:    "AMD Ryzen 9",
			},
		},
		{
			name: "should create CPUMetrics with zero values",
			data: CPUMetrics{
				UsagePercent: 0.0,
				CoreCount:    0,
				ModelName:    "",
			},
			expected: CPUMetrics{
				UsagePercent: 0.0,
				CoreCount:    0,
				ModelName:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.UsagePercent != tt.expected.UsagePercent {
				t.Errorf("UsagePercent = %v, want %v", tt.data.UsagePercent, tt.expected.UsagePercent)
			}
			if tt.data.CoreCount != tt.expected.CoreCount {
				t.Errorf("CoreCount = %v, want %v", tt.data.CoreCount, tt.expected.CoreCount)
			}
			if tt.data.ModelName != tt.expected.ModelName {
				t.Errorf("ModelName = %v, want %v", tt.data.ModelName, tt.expected.ModelName)
			}
		})
	}
}

// TestMemoryMetrics_Struct 测试 MemoryMetrics 结构体
func TestMemoryMetrics_Struct(t *testing.T) {
	tests := []struct {
		name     string
		data     MemoryMetrics
		expected MemoryMetrics
	}{
		{
			name: "should create MemoryMetrics with valid values",
			data: MemoryMetrics{
				TotalGB:      64.0,
				UsedGB:       32.0,
				FreeGB:       32.0,
				UsagePercent: 50.0,
			},
			expected: MemoryMetrics{
				TotalGB:      64.0,
				UsedGB:       32.0,
				FreeGB:       32.0,
				UsagePercent: 50.0,
			},
		},
		{
			name: "should create MemoryMetrics with zero values",
			data: MemoryMetrics{
				TotalGB:      0.0,
				UsedGB:       0.0,
				FreeGB:       0.0,
				UsagePercent: 0.0,
			},
			expected: MemoryMetrics{
				TotalGB:      0.0,
				UsedGB:       0.0,
				FreeGB:       0.0,
				UsagePercent: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.TotalGB != tt.expected.TotalGB {
				t.Errorf("TotalGB = %v, want %v", tt.data.TotalGB, tt.expected.TotalGB)
			}
			if tt.data.UsedGB != tt.expected.UsedGB {
				t.Errorf("UsedGB = %v, want %v", tt.data.UsedGB, tt.expected.UsedGB)
			}
			if tt.data.FreeGB != tt.expected.FreeGB {
				t.Errorf("FreeGB = %v, want %v", tt.data.FreeGB, tt.expected.FreeGB)
			}
			if tt.data.UsagePercent != tt.expected.UsagePercent {
				t.Errorf("UsagePercent = %v, want %v", tt.data.UsagePercent, tt.expected.UsagePercent)
			}
		})
	}
}

// TestDiskMetrics_Struct 测试 DiskMetrics 结构体
func TestDiskMetrics_Struct(t *testing.T) {
	tests := []struct {
		name     string
		data     DiskMetrics
		expected DiskMetrics
	}{
		{
			name: "should create DiskMetrics with valid values",
			data: DiskMetrics{
				MountPoint:   "/data",
				Device:       "/dev/sdb1",
				TotalGB:      1000.0,
				UsedGB:       500.0,
				FreeGB:       500.0,
				UsagePercent: 50.0,
			},
			expected: DiskMetrics{
				MountPoint:   "/data",
				Device:       "/dev/sdb1",
				TotalGB:      1000.0,
				UsedGB:       500.0,
				FreeGB:       500.0,
				UsagePercent: 50.0,
			},
		},
		{
			name: "should create DiskMetrics with empty strings",
			data: DiskMetrics{
				MountPoint:   "",
				Device:       "",
				TotalGB:      0.0,
				UsedGB:       0.0,
				FreeGB:       0.0,
				UsagePercent: 0.0,
			},
			expected: DiskMetrics{
				MountPoint:   "",
				Device:       "",
				TotalGB:      0.0,
				UsedGB:       0.0,
				FreeGB:       0.0,
				UsagePercent: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.MountPoint != tt.expected.MountPoint {
				t.Errorf("MountPoint = %v, want %v", tt.data.MountPoint, tt.expected.MountPoint)
			}
			if tt.data.Device != tt.expected.Device {
				t.Errorf("Device = %v, want %v", tt.data.Device, tt.expected.Device)
			}
			if tt.data.TotalGB != tt.expected.TotalGB {
				t.Errorf("TotalGB = %v, want %v", tt.data.TotalGB, tt.expected.TotalGB)
			}
			if tt.data.UsedGB != tt.expected.UsedGB {
				t.Errorf("UsedGB = %v, want %v", tt.data.UsedGB, tt.expected.UsedGB)
			}
			if tt.data.FreeGB != tt.expected.FreeGB {
				t.Errorf("FreeGB = %v, want %v", tt.data.FreeGB, tt.expected.FreeGB)
			}
			if tt.data.UsagePercent != tt.expected.UsagePercent {
				t.Errorf("UsagePercent = %v, want %v", tt.data.UsagePercent, tt.expected.UsagePercent)
			}
		})
	}
}

// TestNetworkMetrics_Struct 测试 NetworkMetrics 结构体
func TestNetworkMetrics_Struct(t *testing.T) {
	tests := []struct {
		name     string
		data     NetworkMetrics
		expected NetworkMetrics
	}{
		{
			name: "should create NetworkMetrics with valid values",
			data: NetworkMetrics{
				InterfaceName: "eth0",
				BytesSentMB:   1024.5,
				BytesRecvMB:   2048.5,
				PacketsSent:   100000,
				PacketsRecv:   200000,
			},
			expected: NetworkMetrics{
				InterfaceName: "eth0",
				BytesSentMB:   1024.5,
				BytesRecvMB:   2048.5,
				PacketsSent:   100000,
				PacketsRecv:   200000,
			},
		},
		{
			name: "should create NetworkMetrics with zero values",
			data: NetworkMetrics{
				InterfaceName: "",
				BytesSentMB:   0.0,
				BytesRecvMB:   0.0,
				PacketsSent:   0,
				PacketsRecv:   0,
			},
			expected: NetworkMetrics{
				InterfaceName: "",
				BytesSentMB:   0.0,
				BytesRecvMB:   0.0,
				PacketsSent:   0,
				PacketsRecv:   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.InterfaceName != tt.expected.InterfaceName {
				t.Errorf("InterfaceName = %v, want %v", tt.data.InterfaceName, tt.expected.InterfaceName)
			}
			if tt.data.BytesSentMB != tt.expected.BytesSentMB {
				t.Errorf("BytesSentMB = %v, want %v", tt.data.BytesSentMB, tt.expected.BytesSentMB)
			}
			if tt.data.BytesRecvMB != tt.expected.BytesRecvMB {
				t.Errorf("BytesRecvMB = %v, want %v", tt.data.BytesRecvMB, tt.expected.BytesRecvMB)
			}
			if tt.data.PacketsSent != tt.expected.PacketsSent {
				t.Errorf("PacketsSent = %v, want %v", tt.data.PacketsSent, tt.expected.PacketsSent)
			}
			if tt.data.PacketsRecv != tt.expected.PacketsRecv {
				t.Errorf("PacketsRecv = %v, want %v", tt.data.PacketsRecv, tt.expected.PacketsRecv)
			}
		})
	}
}

// TestLoadMetrics_Struct 测试 LoadMetrics 结构体
func TestLoadMetrics_Struct(t *testing.T) {
	tests := []struct {
		name     string
		data     LoadMetrics
		expected LoadMetrics
	}{
		{
			name: "should create LoadMetrics with valid values",
			data: LoadMetrics{
				Load1:  2.5,
				Load5:  2.0,
				Load15: 1.5,
			},
			expected: LoadMetrics{
				Load1:  2.5,
				Load5:  2.0,
				Load15: 1.5,
			},
		},
		{
			name: "should create LoadMetrics with zero values",
			data: LoadMetrics{
				Load1:  0.0,
				Load5:  0.0,
				Load15: 0.0,
			},
			expected: LoadMetrics{
				Load1:  0.0,
				Load5:  0.0,
				Load15: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.Load1 != tt.expected.Load1 {
				t.Errorf("Load1 = %v, want %v", tt.data.Load1, tt.expected.Load1)
			}
			if tt.data.Load5 != tt.expected.Load5 {
				t.Errorf("Load5 = %v, want %v", tt.data.Load5, tt.expected.Load5)
			}
			if tt.data.Load15 != tt.expected.Load15 {
				t.Errorf("Load15 = %v, want %v", tt.data.Load15, tt.expected.Load15)
			}
		})
	}
}

// TestLLMNodeMetric_TableName 测试 LLMNodeMetric 表名
func TestLLMNodeMetric_TableName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "should return llm_node_metrics",
			expected: "llm_node_metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lm := LLMNodeMetric{}
			got := lm.TableName()
			if got != tt.expected {
				t.Errorf("TableName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLLMNodeMetric_Constants 测试 NodeStatus 常量
func TestLLMNodeMetric_Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{
			name:     "NodeStatusOnline should be online",
			constant: NodeStatusOnline,
			expected: "online",
		},
		{
			name:     "NodeStatusOffline should be offline",
			constant: NodeStatusOffline,
			expected: "offline",
		},
		{
			name:     "NodeStatusBusy should be busy",
			constant: NodeStatusBusy,
			expected: "busy",
		},
		{
			name:     "NodeStatusError should be error",
			constant: NodeStatusError,
			expected: "error",
		},
		{
			name:     "NodeStatusIdle should be idle",
			constant: NodeStatusIdle,
			expected: "idle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("constant = %v, want %v", tt.constant, tt.expected)
			}
		})
	}
}

// TestLLMNodeMetric_Struct 测试 LLMNodeMetric 结构体创建和字段赋值
func TestLLMNodeMetric_Struct(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		data     LLMNodeMetric
		expected LLMNodeMetric
	}{
		{
			name: "should create LLMNodeMetric with all fields",
			data: LLMNodeMetric{
				ID:                1,
				NodeID:            "node-001",
				NodeName:          "GPU Node 1",
				Status:            NodeStatusOnline,
				GPUInfo:           `[{"index":0,"name":"NVIDIA A100"}]`,
				GPUCount:          4,
				GPUTotalMemoryGB:  320.0,
				GPUUsedMemoryGB:   160.0,
				GPUUtilization:    75.5,
				CPUCores:          32,
				CPUUsagePercent:   45.0,
				MemoryTotalGB:     256.0,
				MemoryUsedGB:      128.0,
				RequestsPerMin:    100,
				AvgResponseTimeMs: 150.5,
				ActiveRequests:    50,
				QueuedRequests:    10,
				LoadedModels:      `["model-1","model-2"]`,
				ModelCount:        2,
				Version:           "v1.2.3",
				Labels:            `{"region":"us-west"}`,
				ReportedAt:        now,
				CreatedAt:         now,
			},
			expected: LLMNodeMetric{
				ID:                1,
				NodeID:            "node-001",
				NodeName:          "GPU Node 1",
				Status:            "online",
				GPUInfo:           `[{"index":0,"name":"NVIDIA A100"}]`,
				GPUCount:          4,
				GPUTotalMemoryGB:  320.0,
				GPUUsedMemoryGB:   160.0,
				GPUUtilization:    75.5,
				CPUCores:          32,
				CPUUsagePercent:   45.0,
				MemoryTotalGB:     256.0,
				MemoryUsedGB:      128.0,
				RequestsPerMin:    100,
				AvgResponseTimeMs: 150.5,
				ActiveRequests:    50,
				QueuedRequests:    10,
				LoadedModels:      `["model-1","model-2"]`,
				ModelCount:        2,
				Version:           "v1.2.3",
				Labels:            `{"region":"us-west"}`,
				ReportedAt:        now,
				CreatedAt:         now,
			},
		},
		{
			name: "should create LLMNodeMetric with offline status",
			data: LLMNodeMetric{
				ID:       2,
				NodeID:   "node-002",
				NodeName: "GPU Node 2",
				Status:   NodeStatusOffline,
			},
			expected: LLMNodeMetric{
				ID:       2,
				NodeID:   "node-002",
				NodeName: "GPU Node 2",
				Status:   "offline",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", tt.data.ID, tt.expected.ID)
			}
			if tt.data.NodeID != tt.expected.NodeID {
				t.Errorf("NodeID = %v, want %v", tt.data.NodeID, tt.expected.NodeID)
			}
			if tt.data.NodeName != tt.expected.NodeName {
				t.Errorf("NodeName = %v, want %v", tt.data.NodeName, tt.expected.NodeName)
			}
			if tt.data.Status != tt.expected.Status {
				t.Errorf("Status = %v, want %v", tt.data.Status, tt.expected.Status)
			}
			if tt.data.GPUCount != tt.expected.GPUCount {
				t.Errorf("GPUCount = %v, want %v", tt.data.GPUCount, tt.expected.GPUCount)
			}
			if tt.data.GPUUtilization != tt.expected.GPUUtilization {
				t.Errorf("GPUUtilization = %v, want %v", tt.data.GPUUtilization, tt.expected.GPUUtilization)
			}
		})
	}
}

// TestGPUInfo_Struct 测试 GPUInfo 结构体
func TestGPUInfo_Struct(t *testing.T) {
	tests := []struct {
		name     string
		data     GPUInfo
		expected GPUInfo
	}{
		{
			name: "should create GPUInfo with valid values",
			data: GPUInfo{
				Index:         0,
				Name:          "NVIDIA A100",
				TotalMemoryGB: 80.0,
				UsedMemoryGB:  40.0,
				Temperature:   65,
				Utilization:   75.5,
			},
			expected: GPUInfo{
				Index:         0,
				Name:          "NVIDIA A100",
				TotalMemoryGB: 80.0,
				UsedMemoryGB:  40.0,
				Temperature:   65,
				Utilization:   75.5,
			},
		},
		{
			name: "should create GPUInfo with zero values",
			data: GPUInfo{
				Index:         1,
				Name:          "",
				TotalMemoryGB: 0.0,
				UsedMemoryGB:  0.0,
				Temperature:   0,
				Utilization:   0.0,
			},
			expected: GPUInfo{
				Index:         1,
				Name:          "",
				TotalMemoryGB: 0.0,
				UsedMemoryGB:  0.0,
				Temperature:   0,
				Utilization:   0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.Index != tt.expected.Index {
				t.Errorf("Index = %v, want %v", tt.data.Index, tt.expected.Index)
			}
			if tt.data.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", tt.data.Name, tt.expected.Name)
			}
			if tt.data.TotalMemoryGB != tt.expected.TotalMemoryGB {
				t.Errorf("TotalMemoryGB = %v, want %v", tt.data.TotalMemoryGB, tt.expected.TotalMemoryGB)
			}
			if tt.data.UsedMemoryGB != tt.expected.UsedMemoryGB {
				t.Errorf("UsedMemoryGB = %v, want %v", tt.data.UsedMemoryGB, tt.expected.UsedMemoryGB)
			}
			if tt.data.Temperature != tt.expected.Temperature {
				t.Errorf("Temperature = %v, want %v", tt.data.Temperature, tt.expected.Temperature)
			}
			if tt.data.Utilization != tt.expected.Utilization {
				t.Errorf("Utilization = %v, want %v", tt.data.Utilization, tt.expected.Utilization)
			}
		})
	}
}

// TestLoadedModelInfo_Struct 测试 LoadedModelInfo 结构体
func TestLoadedModelInfo_Struct(t *testing.T) {
	now := time.Now()
	loadedAt := now.Add(-time.Hour)

	tests := []struct {
		name     string
		data     LoadedModelInfo
		expected LoadedModelInfo
	}{
		{
			name: "should create LoadedModelInfo with valid values",
			data: LoadedModelInfo{
				ModelID:      "model-001",
				ModelName:    "GPT-4",
				LoadedAt:     loadedAt,
				RequestCount: 1000,
			},
			expected: LoadedModelInfo{
				ModelID:      "model-001",
				ModelName:    "GPT-4",
				LoadedAt:     loadedAt,
				RequestCount: 1000,
			},
		},
		{
			name: "should create LoadedModelInfo with zero values",
			data: LoadedModelInfo{
				ModelID:      "",
				ModelName:    "",
				LoadedAt:     now,
				RequestCount: 0,
			},
			expected: LoadedModelInfo{
				ModelID:      "",
				ModelName:    "",
				LoadedAt:     now,
				RequestCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.ModelID != tt.expected.ModelID {
				t.Errorf("ModelID = %v, want %v", tt.data.ModelID, tt.expected.ModelID)
			}
			if tt.data.ModelName != tt.expected.ModelName {
				t.Errorf("ModelName = %v, want %v", tt.data.ModelName, tt.expected.ModelName)
			}
			if !tt.data.LoadedAt.Equal(tt.expected.LoadedAt) {
				t.Errorf("LoadedAt = %v, want %v", tt.data.LoadedAt, tt.expected.LoadedAt)
			}
			if tt.data.RequestCount != tt.expected.RequestCount {
				t.Errorf("RequestCount = %v, want %v", tt.data.RequestCount, tt.expected.RequestCount)
			}
		})
	}
}

// TestLLMNodeSummary_Struct 测试 LLMNodeSummary 结构体
func TestLLMNodeSummary_Struct(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		data     LLMNodeSummary
		expected LLMNodeSummary
	}{
		{
			name: "should create LLMNodeSummary with all fields",
			data: LLMNodeSummary{
				NodeID:             "node-001",
				NodeName:           "GPU Node 1",
				Status:             NodeStatusOnline,
				GPUUtilization:     75.5,
				GPUTotalMemoryGB:   320.0,
				GPUUsedMemoryGB:    160.0,
				CPUUsagePercent:    45.0,
				MemoryUsagePercent: 50.0,
				RequestsPerMin:     100,
				AvgResponseTimeMs:  150.5,
				ActiveRequests:     50,
				ModelCount:         2,
				LoadedModels:       []string{"model-1", "model-2"},
				Version:            "v1.2.3",
				LastSeenAt:         now,
			},
			expected: LLMNodeSummary{
				NodeID:             "node-001",
				NodeName:           "GPU Node 1",
				Status:             "online",
				GPUUtilization:     75.5,
				GPUTotalMemoryGB:   320.0,
				GPUUsedMemoryGB:    160.0,
				CPUUsagePercent:    45.0,
				MemoryUsagePercent: 50.0,
				RequestsPerMin:     100,
				AvgResponseTimeMs:  150.5,
				ActiveRequests:     50,
				ModelCount:         2,
				LoadedModels:       []string{"model-1", "model-2"},
				Version:            "v1.2.3",
				LastSeenAt:         now,
			},
		},
		{
			name: "should create LLMNodeSummary with empty slices",
			data: LLMNodeSummary{
				NodeID:       "node-002",
				NodeName:     "GPU Node 2",
				Status:       NodeStatusOffline,
				LoadedModels: []string{},
				LastSeenAt:   now,
			},
			expected: LLMNodeSummary{
				NodeID:       "node-002",
				NodeName:     "GPU Node 2",
				Status:       "offline",
				LoadedModels: []string{},
				LastSeenAt:   now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.NodeID != tt.expected.NodeID {
				t.Errorf("NodeID = %v, want %v", tt.data.NodeID, tt.expected.NodeID)
			}
			if tt.data.NodeName != tt.expected.NodeName {
				t.Errorf("NodeName = %v, want %v", tt.data.NodeName, tt.expected.NodeName)
			}
			if tt.data.Status != tt.expected.Status {
				t.Errorf("Status = %v, want %v", tt.data.Status, tt.expected.Status)
			}
			if tt.data.GPUUtilization != tt.expected.GPUUtilization {
				t.Errorf("GPUUtilization = %v, want %v", tt.data.GPUUtilization, tt.expected.GPUUtilization)
			}
			if tt.data.GPUTotalMemoryGB != tt.expected.GPUTotalMemoryGB {
				t.Errorf("GPUTotalMemoryGB = %v, want %v", tt.data.GPUTotalMemoryGB, tt.expected.GPUTotalMemoryGB)
			}
			if tt.data.GPUUsedMemoryGB != tt.expected.GPUUsedMemoryGB {
				t.Errorf("GPUUsedMemoryGB = %v, want %v", tt.data.GPUUsedMemoryGB, tt.expected.GPUUsedMemoryGB)
			}
			if tt.data.CPUUsagePercent != tt.expected.CPUUsagePercent {
				t.Errorf("CPUUsagePercent = %v, want %v", tt.data.CPUUsagePercent, tt.expected.CPUUsagePercent)
			}
			if tt.data.MemoryUsagePercent != tt.expected.MemoryUsagePercent {
				t.Errorf("MemoryUsagePercent = %v, want %v", tt.data.MemoryUsagePercent, tt.expected.MemoryUsagePercent)
			}
			if tt.data.RequestsPerMin != tt.expected.RequestsPerMin {
				t.Errorf("RequestsPerMin = %v, want %v", tt.data.RequestsPerMin, tt.expected.RequestsPerMin)
			}
			if tt.data.AvgResponseTimeMs != tt.expected.AvgResponseTimeMs {
				t.Errorf("AvgResponseTimeMs = %v, want %v", tt.data.AvgResponseTimeMs, tt.expected.AvgResponseTimeMs)
			}
			if tt.data.ActiveRequests != tt.expected.ActiveRequests {
				t.Errorf("ActiveRequests = %v, want %v", tt.data.ActiveRequests, tt.expected.ActiveRequests)
			}
			if tt.data.ModelCount != tt.expected.ModelCount {
				t.Errorf("ModelCount = %v, want %v", tt.data.ModelCount, tt.expected.ModelCount)
			}
			if len(tt.data.LoadedModels) != len(tt.expected.LoadedModels) {
				t.Errorf("LoadedModels length = %v, want %v", len(tt.data.LoadedModels), len(tt.expected.LoadedModels))
			}
			if tt.data.Version != tt.expected.Version {
				t.Errorf("Version = %v, want %v", tt.data.Version, tt.expected.Version)
			}
			if !tt.data.LastSeenAt.Equal(tt.expected.LastSeenAt) {
				t.Errorf("LastSeenAt = %v, want %v", tt.data.LastSeenAt, tt.expected.LastSeenAt)
			}
		})
	}
}
