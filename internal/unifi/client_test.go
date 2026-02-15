package unifi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/builtbyrobben/unifi-cli/internal/api"
)

func newTestClient(server *httptest.Server) *Client {
	return &Client{
		Client: api.NewClient("test-key",
			api.WithBaseURL(server.URL),
			api.WithUserAgent("unifi-cli/test"),
		),
	}
}

func TestHosts_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/hosts" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[[]Host]{
			Data: []Host{
				{ID: "host-1", Type: "console", IPAddress: "192.168.1.1"},
				{ID: "host-2", Type: "console", IPAddress: "10.0.0.1"},
			},
			HTTPStatusCode: 200,
			TraceID:        "trace-123",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Hosts().List(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(result.Data))
	}

	if result.Data[0].ID != "host-1" {
		t.Errorf("expected first host ID 'host-1', got %q", result.Data[0].ID)
	}
}

func TestHosts_List_WithPageSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("pageSize") != "5" {
			t.Errorf("expected pageSize=5, got %s", r.URL.Query().Get("pageSize"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[[]Host]{Data: []Host{}})
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Hosts().List(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHosts_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/hosts/host-123" {
			t.Errorf("expected path /v1/hosts/host-123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[Host]{
			Data: Host{
				ID:        "host-123",
				Type:      "console",
				IPAddress: "192.168.1.1",
				Owner:     true,
			},
			HTTPStatusCode: 200,
			TraceID:        "trace-456",
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Hosts().Get(context.Background(), "host-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Data.ID != "host-123" {
		t.Errorf("expected ID 'host-123', got %q", result.Data.ID)
	}

	if !result.Data.Owner {
		t.Error("expected Owner to be true")
	}
}

func TestHosts_Get_EmptyID(t *testing.T) {
	client := &Client{Client: api.NewClient("key")}

	_, err := client.Hosts().Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestSites_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sites" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[[]Site]{
			Data: []Site{
				{
					SiteID: "site-1",
					HostID: "host-1",
					Meta: SiteMeta{
						Name:     "Home",
						Timezone: "America/Chicago",
					},
					IsOwner: true,
				},
				{
					SiteID: "site-2",
					HostID: "host-1",
					Meta: SiteMeta{
						Name:     "Office",
						Timezone: "America/New_York",
					},
				},
			},
			HTTPStatusCode: 200,
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Sites().List(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 sites, got %d", len(result.Data))
	}

	if result.Data[0].Meta.Name != "Home" {
		t.Errorf("expected first site name 'Home', got %q", result.Data[0].Meta.Name)
	}

	if !result.Data[0].IsOwner {
		t.Error("expected first site IsOwner to be true")
	}
}

func TestDevices_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/devices" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[[]DeviceHost]{
			Data: []DeviceHost{
				{
					HostID:   "host-1",
					HostName: "Home Console",
					Devices: []Device{
						{ID: "dev-1", Name: "Switch", Model: "USW-24", Status: "online", IP: "192.168.1.2"},
						{ID: "dev-2", Name: "AP Pro", Model: "U6-Pro", Status: "online", IP: "192.168.1.3"},
					},
				},
			},
			HTTPStatusCode: 200,
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.Devices().List(context.Background(), "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 device host, got %d", len(result.Data))
	}

	if len(result.Data[0].Devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(result.Data[0].Devices))
	}

	if result.Data[0].Devices[0].Name != "Switch" {
		t.Errorf("expected first device name 'Switch', got %q", result.Data[0].Devices[0].Name)
	}
}

func TestDevices_List_WithHostFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("hostIds[]") != "host-1" {
			t.Errorf("expected hostIds[]=host-1, got %s", r.URL.Query().Get("hostIds[]"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[[]DeviceHost]{Data: []DeviceHost{}})
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.Devices().List(context.Background(), "host-1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestISPMetrics_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/isp-metrics/5m" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.URL.Query().Get("duration") != "24h" {
			t.Errorf("expected duration=24h, got %s", r.URL.Query().Get("duration"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[[]ISPMetricEntry]{
			Data: []ISPMetricEntry{
				{
					MetricType: "5m",
					HostID:     "host-1",
					SiteID:     "site-1",
					Periods: []ISPPeriod{
						{
							MetricTime: "2026-02-15T00:00:00Z",
							Data: &ISPPeriodData{
								WAN: &WANMetrics{
									AvgLatency:   12.5,
									DownloadKbps: 500000,
									UploadKbps:   100000,
									ISPName:      "Test ISP",
								},
							},
						},
					},
				},
			},
			HTTPStatusCode: 200,
		})
	}))
	defer server.Close()

	client := newTestClient(server)

	result, err := client.ISPMetrics().Get(context.Background(), "5m", "24h", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 metric entry, got %d", len(result.Data))
	}

	if result.Data[0].MetricType != "5m" {
		t.Errorf("expected metric type '5m', got %q", result.Data[0].MetricType)
	}

	if len(result.Data[0].Periods) != 1 {
		t.Fatalf("expected 1 period, got %d", len(result.Data[0].Periods))
	}

	wan := result.Data[0].Periods[0].Data.WAN
	if wan.AvgLatency != 12.5 {
		t.Errorf("expected avg latency 12.5, got %f", wan.AvgLatency)
	}

	if wan.ISPName != "Test ISP" {
		t.Errorf("expected ISP name 'Test ISP', got %q", wan.ISPName)
	}
}

func TestISPMetrics_Get_InvalidType(t *testing.T) {
	client := &Client{Client: api.NewClient("key")}

	_, err := client.ISPMetrics().Get(context.Background(), "invalid", "", "", "")
	if err == nil {
		t.Fatal("expected error for invalid metric type")
	}
}

func TestISPMetrics_Get_WithTimestamps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("beginTimestamp") != "1707955200" {
			t.Errorf("expected beginTimestamp=1707955200, got %s", r.URL.Query().Get("beginTimestamp"))
		}

		if r.URL.Query().Get("endTimestamp") != "1708041600" {
			t.Errorf("expected endTimestamp=1708041600, got %s", r.URL.Query().Get("endTimestamp"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[[]ISPMetricEntry]{Data: []ISPMetricEntry{}})
	}))
	defer server.Close()

	client := newTestClient(server)

	_, err := client.ISPMetrics().Get(context.Background(), "1h", "", "1707955200", "1708041600")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
