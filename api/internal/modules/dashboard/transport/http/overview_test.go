package http

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
	"sechelper-auth-template/api/internal/modules/dashboard/application"
)

type fixedResourceSnapshotReader struct {
	value application.ResourceMetrics
	ready bool
}

func (r fixedResourceSnapshotReader) Snapshot() (application.ResourceMetrics, bool) {
	return r.value, r.ready
}

func TestOverviewResponseDataIncludesResourceSnapshot(t *testing.T) {
	cpu := 18.5
	memoryPercent := 62.25
	memoryUsed := uint64(625)
	memoryTotal := uint64(1000)
	diskPercent := 75.0
	diskUsed := uint64(750)
	diskTotal := uint64(1000)
	value := application.Overview{Resources: &application.ResourceMetrics{
		SampledAt:        time.Date(2026, time.September, 24, 10, 10, 1, 0, time.UTC),
		CPUPercent:       &cpu,
		MemoryPercent:    &memoryPercent,
		MemoryUsedBytes:  &memoryUsed,
		MemoryTotalBytes: &memoryTotal,
		DiskPercent:      &diskPercent,
		DiskUsedBytes:    &diskUsed,
		DiskTotalBytes:   &diskTotal,
		Goroutines:       23,
	}}

	body, err := json.Marshal(overviewResponseData(value))
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Resources struct {
			SampledAt        time.Time `json:"sampledAt"`
			CPUPercent       float64   `json:"cpuPercent"`
			MemoryPercent    float64   `json:"memoryPercent"`
			MemoryUsedBytes  uint64    `json:"memoryUsedBytes"`
			MemoryTotalBytes uint64    `json:"memoryTotalBytes"`
			DiskPercent      float64   `json:"diskPercent"`
			DiskUsedBytes    uint64    `json:"diskUsedBytes"`
			DiskTotalBytes   uint64    `json:"diskTotalBytes"`
			Goroutines       int       `json:"goroutines"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	got := decoded.Resources
	if got.SampledAt.Format(time.RFC3339) != "2026-09-24T10:10:01Z" || got.CPUPercent != cpu || got.MemoryPercent != memoryPercent || got.MemoryUsedBytes != memoryUsed || got.MemoryTotalBytes != memoryTotal || got.DiskPercent != diskPercent || got.DiskUsedBytes != diskUsed || got.DiskTotalBytes != diskTotal || got.Goroutines != 23 {
		t.Fatalf("unexpected resource response: %+v", got)
	}
}

func TestOverviewResponseDataOmitsResourcesBeforeFirstSample(t *testing.T) {
	body, err := json.Marshal(overviewResponseData(application.Overview{}))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded["resources"]; ok {
		t.Fatalf("resources should be omitted before the first sample: %s", body)
	}
}

func TestResourceStreamPushesLatestSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cpu := 18.5
	snapshot := application.ResourceMetrics{SampledAt: time.Date(2026, time.September, 24, 10, 10, 1, 0, time.UTC), CPUPercent: &cpu, Goroutines: 23}
	service := application.NewService(nil, application.AppInfo{}, nil, nil, fixedResourceSnapshotReader{value: snapshot, ready: true})
	server := httptest.NewServer(ginEngineForHandler(NewHandler(service)))
	defer server.Close()

	ws, err := websocket.Dial(strings.Replace(server.URL, "http://", "ws://", 1)+"/resources/ws", "", "http://example.test")
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	var message struct {
		Data application.ResourceMetrics `json:"data"`
	}
	if err := websocket.JSON.Receive(ws, &message); err != nil {
		t.Fatal(err)
	}
	if !message.Data.SampledAt.Equal(snapshot.SampledAt) || message.Data.CPUPercent == nil || *message.Data.CPUPercent != cpu || message.Data.Goroutines != 23 {
		t.Fatalf("unexpected resource message: %+v", message.Data)
	}
}

func ginEngineForHandler(handler *Handler) *gin.Engine {
	router := gin.New()
	router.GET("/resources/ws", handler.ResourceStream)
	return router
}
