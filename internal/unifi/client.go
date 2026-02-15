package unifi

import (
	"context"
	"errors"
	"fmt"

	"github.com/builtbyrobben/unifi-cli/internal/api"
)

var (
	errIDRequired       = errors.New("id is required")
	errMetricTypeInvald = errors.New("metric type must be '5m' or '1h'")
)

const defaultBaseURL = "https://api.ui.com"

// Client wraps the API client with UniFi Site Manager methods.
type Client struct {
	*api.Client
}

// NewClient creates a new UniFi Site Manager API client.
func NewClient(apiKey string) *Client {
	return &Client{
		Client: api.NewClient(apiKey,
			api.WithBaseURL(defaultBaseURL),
			api.WithUserAgent("unifi-cli/1.0"),
		),
	}
}

// Response is the standard Site Manager API response envelope.
type Response[T any] struct {
	Data           T      `json:"data"`
	HTTPStatusCode int    `json:"http_status_code"`
	TraceID        string `json:"trace_id"`
	NextToken      string `json:"next_token,omitempty"`
}

// Host represents a UniFi host (console/controller).
type Host struct {
	ID                        string    `json:"id"`
	HardwareID                string    `json:"hardware_id,omitempty"`
	Type                      string    `json:"type,omitempty"`
	IPAddress                 string    `json:"ip_address,omitempty"`
	Owner                     bool      `json:"owner,omitempty"`
	IsBlocked                 bool      `json:"is_blocked,omitempty"`
	RegistrationTime          string    `json:"registration_time,omitempty"`
	LastConnectionStateChange string    `json:"last_connection_state_change,omitempty"`
	LatestBackupTime          string    `json:"latest_backup_time,omitempty"`
	UserData                  *UserData `json:"user_data,omitempty"`
	ReportedState             any       `json:"reported_state,omitempty"`
}

// UserData contains user-defined metadata for a host.
type UserData struct {
	Name string `json:"name,omitempty"`
}

// Site represents a UniFi site.
type Site struct {
	SiteID     string          `json:"site_id"`
	HostID     string          `json:"host_id,omitempty"`
	Meta       SiteMeta        `json:"meta"`
	Statistics *SiteStatistics `json:"statistics,omitempty"`
	Permission string          `json:"permission,omitempty"`
	IsOwner    bool            `json:"is_owner,omitempty"`
}

// SiteMeta contains site metadata.
type SiteMeta struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	GatewayMAC  string `json:"gateway_mac,omitempty"`
}

// SiteStatistics contains site statistics.
type SiteStatistics struct {
	Counts      *SiteCounts      `json:"counts,omitempty"`
	Performance *SitePerformance `json:"performance,omitempty"`
}

// SiteCounts contains device and client counts.
type SiteCounts struct {
	TotalDevice    int `json:"total_device,omitempty"`
	ActiveDevice   int `json:"active_device,omitempty"`
	InactiveDevice int `json:"inactive_device,omitempty"`
	TotalClient    int `json:"total_client,omitempty"`
	WiredClient    int `json:"wired_client,omitempty"`
	WirelessClient int `json:"wireless_client,omitempty"`
}

// SitePerformance contains site performance metrics.
type SitePerformance struct {
	WanLatencyAvg float64 `json:"wan_latency_avg,omitempty"`
	WanUptimePct  float64 `json:"wan_uptime_pct,omitempty"`
}

// DeviceHost represents a host and its devices from the devices endpoint.
type DeviceHost struct {
	HostID   string   `json:"host_id"`
	HostName string   `json:"host_name,omitempty"`
	Devices  []Device `json:"devices"`
	UpdateAt string   `json:"updated_at,omitempty"`
}

// Device represents a UniFi network device.
type Device struct {
	ID             string `json:"id"`
	MAC            string `json:"mac,omitempty"`
	Name           string `json:"name,omitempty"`
	Model          string `json:"model,omitempty"`
	Shortname      string `json:"shortname,omitempty"`
	IP             string `json:"ip,omitempty"`
	ProductLine    string `json:"product_line,omitempty"`
	Status         string `json:"status,omitempty"`
	Version        string `json:"version,omitempty"`
	FirmwareStatus string `json:"firmware_status,omitempty"`
	UpdateAvail    string `json:"update_available,omitempty"`
	IsConsole      bool   `json:"is_console,omitempty"`
	IsManaged      bool   `json:"is_managed,omitempty"`
	StartupTime    string `json:"startup_time,omitempty"`
	AdoptionTime   string `json:"adoption_time,omitempty"`
	Note           string `json:"note,omitempty"`
}

// ISPMetricEntry represents an ISP metric entry from the metrics endpoint.
type ISPMetricEntry struct {
	MetricType string      `json:"metric_type,omitempty"`
	Periods    []ISPPeriod `json:"periods,omitempty"`
	HostID     string      `json:"host_id,omitempty"`
	SiteID     string      `json:"site_id,omitempty"`
}

// ISPPeriod represents a single metric period.
type ISPPeriod struct {
	MetricTime string         `json:"metric_time,omitempty"`
	Data       *ISPPeriodData `json:"data,omitempty"`
}

// ISPPeriodData contains WAN metrics data.
type ISPPeriodData struct {
	WAN *WANMetrics `json:"wan,omitempty"`
}

// WANMetrics contains WAN performance metrics.
type WANMetrics struct {
	AvgLatency   float64 `json:"avg_latency,omitempty"`
	DownloadKbps float64 `json:"download_kbps,omitempty"`
	UploadKbps   float64 `json:"upload_kbps,omitempty"`
	PacketLoss   float64 `json:"packet_loss,omitempty"`
	MaxLatency   float64 `json:"max_latency,omitempty"`
	Uptime       float64 `json:"uptime,omitempty"`
	Downtime     float64 `json:"downtime,omitempty"`
	ISPASN       int     `json:"isp_asn,omitempty"`
	ISPName      string  `json:"isp_name,omitempty"`
}

// Hosts returns the hosts service.
func (c *Client) Hosts() *HostsService {
	return &HostsService{client: c}
}

// Sites returns the sites service.
func (c *Client) Sites() *SitesService {
	return &SitesService{client: c}
}

// Devices returns the devices service.
func (c *Client) Devices() *DevicesService {
	return &DevicesService{client: c}
}

// ISPMetrics returns the ISP metrics service.
func (c *Client) ISPMetrics() *ISPMetricsService {
	return &ISPMetricsService{client: c}
}

// HostsService handles host operations.
type HostsService struct {
	client *Client
}

// List returns all hosts with optional pagination.
func (s *HostsService) List(ctx context.Context, pageSize int) (*Response[[]Host], error) {
	path := "/v1/hosts"

	if pageSize > 0 {
		path = fmt.Sprintf("/v1/hosts?pageSize=%d", pageSize)
	}

	var result Response[[]Host]
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list hosts: %w", err)
	}

	return &result, nil
}

// Get returns a host by ID.
func (s *HostsService) Get(ctx context.Context, id string) (*Response[Host], error) {
	if id == "" {
		return nil, errIDRequired
	}

	path := fmt.Sprintf("/v1/hosts/%s", id)

	var result Response[Host]
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get host: %w", err)
	}

	return &result, nil
}

// SitesService handles site operations.
type SitesService struct {
	client *Client
}

// List returns all sites with optional pagination.
func (s *SitesService) List(ctx context.Context, pageSize int) (*Response[[]Site], error) {
	path := "/v1/sites"

	if pageSize > 0 {
		path = fmt.Sprintf("/v1/sites?pageSize=%d", pageSize)
	}

	var result Response[[]Site]
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list sites: %w", err)
	}

	return &result, nil
}

// DevicesService handles device operations.
type DevicesService struct {
	client *Client
}

// List returns all devices, optionally filtered by host ID.
func (s *DevicesService) List(ctx context.Context, hostID string, pageSize int) (*Response[[]DeviceHost], error) {
	path := "/v1/devices"
	sep := "?"

	if hostID != "" {
		path += fmt.Sprintf("%shostIds[]=%s", sep, hostID)
		sep = "&"
	}

	if pageSize > 0 {
		path += fmt.Sprintf("%spageSize=%d", sep, pageSize)
	}

	var result Response[[]DeviceHost]
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}

	return &result, nil
}

// ISPMetricsService handles ISP metrics operations.
type ISPMetricsService struct {
	client *Client
}

// Get returns ISP metrics for the given type (5m or 1h).
func (s *ISPMetricsService) Get(ctx context.Context, metricType, duration string, beginTS, endTS string) (*Response[[]ISPMetricEntry], error) {
	if metricType != "5m" && metricType != "1h" {
		return nil, errMetricTypeInvald
	}

	path := fmt.Sprintf("/v1/isp-metrics/%s", metricType)
	sep := "?"

	if duration != "" {
		path += fmt.Sprintf("%sduration=%s", sep, duration)
		sep = "&"
	}

	if beginTS != "" {
		path += fmt.Sprintf("%sbeginTimestamp=%s", sep, beginTS)
		sep = "&"
	}

	if endTS != "" {
		path += fmt.Sprintf("%sendTimestamp=%s", sep, endTS)
	}

	var result Response[[]ISPMetricEntry]
	if err := s.client.Get(ctx, path, &result); err != nil {
		return nil, fmt.Errorf("get isp metrics: %w", err)
	}

	return &result, nil
}
