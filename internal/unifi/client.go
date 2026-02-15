package unifi

import (
	"context"
	"errors"
	"fmt"

	"github.com/builtbyrobben/unifi-cli/internal/api"
)

var (
	errMACRequired = errors.New("MAC address is required")
	errIDRequired  = errors.New("ID is required")
	errNotFound    = errors.New("not found")
)

// Client wraps the API client with UniFi-specific methods.
type Client struct {
	*api.Client
}

// NewClient creates a new UniFi service client.
func NewClient(apiClient *api.Client) *Client {
	return &Client{Client: apiClient}
}

// Response is the standard UniFi API response envelope.
type Response[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

// Meta contains response metadata.
type Meta struct {
	RC  string `json:"rc"`
	Msg string `json:"msg,omitempty"`
}

// Device represents a UniFi network device.
type Device struct {
	MAC             string  `json:"mac"`
	IP              string  `json:"ip,omitempty"`
	Name            string  `json:"name,omitempty"`
	Model           string  `json:"model,omitempty"`
	Type            string  `json:"type,omitempty"`
	Version         string  `json:"version,omitempty"`
	Adopted         bool    `json:"adopted"`
	State           int     `json:"state"`
	Uptime          int64   `json:"uptime,omitempty"`
	LastSeen        int64   `json:"last_seen,omitempty"`
	Upgradable      bool    `json:"upgradable,omitempty"`
	NumSta          int     `json:"num_sta,omitempty"`
	TxBytes         int64   `json:"tx_bytes,omitempty"`
	RxBytes         int64   `json:"rx_bytes,omitempty"`
	SatisfactionAvg float64 `json:"satisfaction,omitempty"`
}

// ClientDevice represents a network client (station).
type ClientDevice struct {
	MAC        string `json:"mac"`
	IP         string `json:"ip,omitempty"`
	Hostname   string `json:"hostname,omitempty"`
	Name       string `json:"name,omitempty"`
	IsWired    bool   `json:"is_wired"`
	Network    string `json:"network,omitempty"`
	NetworkID  string `json:"network_id,omitempty"`
	Uptime     int64  `json:"uptime,omitempty"`
	LastSeen   int64  `json:"last_seen,omitempty"`
	TxBytes    int64  `json:"tx_bytes,omitempty"`
	RxBytes    int64  `json:"rx_bytes,omitempty"`
	Blocked    bool   `json:"blocked,omitempty"`
	IsGuest    bool   `json:"is_guest,omitempty"`
	Noted      bool   `json:"noted,omitempty"`
	Note       string `json:"note,omitempty"`
	DeviceName string `json:"device_name,omitempty"`
}

// Network represents a UniFi network/VLAN configuration.
type Network struct {
	ID                 string `json:"_id"` //nolint:tagliatelle // UniFi API uses _id
	Name               string `json:"name"`
	Purpose            string `json:"purpose,omitempty"`
	IPSubnet           string `json:"ip_subnet,omitempty"`
	VLAN               int    `json:"vlan,omitempty"`
	VLANEnabled        bool   `json:"vlan_enabled,omitempty"`
	DomainName         string `json:"domain_name,omitempty"`
	DHCPDEnabled       bool   `json:"dhcpd_enabled,omitempty"`
	DHCPDStart         string `json:"dhcpd_start,omitempty"`
	DHCPDStop          string `json:"dhcpd_stop,omitempty"`
	NetworkGroup       string `json:"networkgroup,omitempty"`
	IsNAT              bool   `json:"is_nat,omitempty"`
	InternetAccessible bool   `json:"internet_access_enabled,omitempty"`
}

// HealthEntry represents a site health entry.
type HealthEntry struct {
	Subsystem       string  `json:"subsystem"`
	Status          string  `json:"status"`
	NumUser         int     `json:"num_user,omitempty"`
	NumGuest        int     `json:"num_guest,omitempty"`
	NumAdopted      int     `json:"num_adopted,omitempty"`
	NumDisconnected int     `json:"num_disconnected,omitempty"`
	NumPending      int     `json:"num_pending,omitempty"`
	TxBytesR        float64 `json:"tx_bytes-r,omitempty"` //nolint:tagliatelle // UniFi API field name
	RxBytesR        float64 `json:"rx_bytes-r,omitempty"` //nolint:tagliatelle // UniFi API field name
	ISPName         string  `json:"isp_name,omitempty"`
	ISPOrganization string  `json:"isp_organization,omitempty"`
	Latency         int     `json:"latency,omitempty"`
	Uptime          int64   `json:"uptime,omitempty"`
}

// SysInfo represents site system information.
type SysInfo struct {
	Timezone    string   `json:"timezone,omitempty"`
	Version     string   `json:"version,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	Name        string   `json:"name,omitempty"`
	IPAddrs     []string `json:"ip_addrs,omitempty"`
	UpdateAvail bool     `json:"update_available,omitempty"`
	LiveChat    string   `json:"live_chat,omitempty"`
	AutoBackup  bool     `json:"autobackup,omitempty"`
	BuildNumber string   `json:"build,omitempty"`
}

// Devices returns the devices service.
func (c *Client) Devices() *DevicesService {
	return &DevicesService{client: c}
}

// Clients returns the clients service.
func (c *Client) Clients() *ClientsService {
	return &ClientsService{client: c}
}

// Networks returns the networks service.
func (c *Client) Networks() *NetworksService {
	return &NetworksService{client: c}
}

// Stats returns the stats service.
func (c *Client) Stats() *StatsService {
	return &StatsService{client: c}
}

// DevicesService handles device operations.
type DevicesService struct {
	client *Client
}

// List returns all network devices, optionally filtered by type.
func (s *DevicesService) List(ctx context.Context, deviceType string) ([]Device, error) {
	path := s.client.SitePath() + "/stat/device"

	var resp Response[Device]
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("list devices: %w", err)
	}

	if deviceType == "" {
		return resp.Data, nil
	}

	var filtered []Device

	for _, d := range resp.Data {
		if d.Type == deviceType {
			filtered = append(filtered, d)
		}
	}

	return filtered, nil
}

// Get returns a device by MAC address.
func (s *DevicesService) Get(ctx context.Context, mac string) (*Device, error) {
	if mac == "" {
		return nil, errMACRequired
	}

	path := s.client.SitePath() + "/stat/device/" + mac

	var resp Response[Device]
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("get device: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("device %s: %w", mac, errNotFound)
	}

	return &resp.Data[0], nil
}

// Restart restarts a device by MAC address.
func (s *DevicesService) Restart(ctx context.Context, mac string) error {
	if mac == "" {
		return errMACRequired
	}

	path := s.client.SitePath() + "/cmd/devmgr"
	body := map[string]string{
		"cmd": "restart",
		"mac": mac,
	}

	var resp Response[any]
	if err := s.client.Post(ctx, path, body, &resp); err != nil {
		return fmt.Errorf("restart device: %w", err)
	}

	return nil
}

// ClientsService handles client operations.
type ClientsService struct {
	client *Client
}

// ListActive returns currently connected clients, optionally filtered by type.
func (s *ClientsService) ListActive(ctx context.Context, clientType string) ([]ClientDevice, error) {
	path := s.client.SitePath() + "/stat/sta"

	var resp Response[ClientDevice]
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("list active clients: %w", err)
	}

	return filterClients(resp.Data, clientType), nil
}

// ListAll returns all known clients, optionally filtered by type.
func (s *ClientsService) ListAll(ctx context.Context, clientType string) ([]ClientDevice, error) {
	path := s.client.SitePath() + "/rest/user"

	var resp Response[ClientDevice]
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("list all clients: %w", err)
	}

	return filterClients(resp.Data, clientType), nil
}

// Get returns a client by MAC address.
func (s *ClientsService) Get(ctx context.Context, mac string) (*ClientDevice, error) {
	if mac == "" {
		return nil, errMACRequired
	}

	path := s.client.SitePath() + "/stat/sta/" + mac

	var resp Response[ClientDevice]
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("get client: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("client %s: %w", mac, errNotFound)
	}

	return &resp.Data[0], nil
}

// Block blocks a client by MAC address.
func (s *ClientsService) Block(ctx context.Context, mac string) error {
	if mac == "" {
		return errMACRequired
	}

	path := s.client.SitePath() + "/cmd/stamgr"
	body := map[string]string{
		"cmd": "block-sta",
		"mac": mac,
	}

	var resp Response[any]
	if err := s.client.Post(ctx, path, body, &resp); err != nil {
		return fmt.Errorf("block client: %w", err)
	}

	return nil
}

// Unblock unblocks a client by MAC address.
func (s *ClientsService) Unblock(ctx context.Context, mac string) error {
	if mac == "" {
		return errMACRequired
	}

	path := s.client.SitePath() + "/cmd/stamgr"
	body := map[string]string{
		"cmd": "unblock-sta",
		"mac": mac,
	}

	var resp Response[any]
	if err := s.client.Post(ctx, path, body, &resp); err != nil {
		return fmt.Errorf("unblock client: %w", err)
	}

	return nil
}

func filterClients(clients []ClientDevice, clientType string) []ClientDevice {
	if clientType == "" {
		return clients
	}

	var filtered []ClientDevice

	for _, c := range clients {
		switch clientType {
		case "wired":
			if c.IsWired {
				filtered = append(filtered, c)
			}
		case "wireless":
			if !c.IsWired {
				filtered = append(filtered, c)
			}
		}
	}

	return filtered
}

// NetworksService handles network operations.
type NetworksService struct {
	client *Client
}

// List returns all network/VLAN configurations.
func (s *NetworksService) List(ctx context.Context) ([]Network, error) {
	path := s.client.SitePath() + "/rest/networkconf"

	var resp Response[Network]
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("list networks: %w", err)
	}

	return resp.Data, nil
}

// Get returns a specific network configuration by ID.
func (s *NetworksService) Get(ctx context.Context, id string) (*Network, error) {
	if id == "" {
		return nil, errIDRequired
	}

	path := s.client.SitePath() + "/rest/networkconf/" + id

	var resp Response[Network]
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("get network: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("network %s: %w", id, errNotFound)
	}

	return &resp.Data[0], nil
}

// StatsService handles stats operations.
type StatsService struct {
	client *Client
}

// Health returns the site health overview.
func (s *StatsService) Health(ctx context.Context) ([]HealthEntry, error) {
	path := s.client.SitePath() + "/stat/health"

	var resp Response[HealthEntry]
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("get health: %w", err)
	}

	return resp.Data, nil
}

// SysInfo returns site system information.
func (s *StatsService) SysInfo(ctx context.Context) ([]SysInfo, error) {
	path := s.client.SitePath() + "/stat/sysinfo"

	var resp Response[SysInfo]
	if err := s.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("get sysinfo: %w", err)
	}

	return resp.Data, nil
}
