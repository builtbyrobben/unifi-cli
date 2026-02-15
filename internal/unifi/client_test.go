package unifi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/builtbyrobben/unifi-cli/internal/api"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/login", func(w http.ResponseWriter, _ *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "TOKEN", Value: "token", Path: "/"})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/proxy/network/api/s/default/stat/device", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[Device]{
			Data: []Device{
				{MAC: "AA:BB:CC:DD:EE:FF", Name: "Switch", Type: "usw", IP: "192.168.1.2", Adopted: true},
				{MAC: "11:22:33:44:55:66", Name: "AP", Type: "uap", IP: "192.168.1.3", Adopted: true},
			},
			Meta: Meta{RC: "ok"},
		})
	})

	mux.HandleFunc("/proxy/network/api/s/default/stat/device/AA:BB:CC:DD:EE:FF", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[Device]{
			Data: []Device{
				{MAC: "AA:BB:CC:DD:EE:FF", Name: "Switch", Type: "usw", IP: "192.168.1.2", Adopted: true, NumSta: 5},
			},
			Meta: Meta{RC: "ok"},
		})
	})

	mux.HandleFunc("/proxy/network/api/s/default/cmd/devmgr", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[any]{Meta: Meta{RC: "ok"}})
	})

	mux.HandleFunc("/proxy/network/api/s/default/stat/sta", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[ClientDevice]{
			Data: []ClientDevice{
				{MAC: "CC:DD:EE:FF:00:11", Hostname: "laptop", IP: "192.168.1.100", IsWired: false},
				{MAC: "DD:EE:FF:00:11:22", Hostname: "desktop", IP: "192.168.1.101", IsWired: true},
			},
			Meta: Meta{RC: "ok"},
		})
	})

	mux.HandleFunc("/proxy/network/api/s/default/stat/sta/CC:DD:EE:FF:00:11", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[ClientDevice]{
			Data: []ClientDevice{
				{MAC: "CC:DD:EE:FF:00:11", Hostname: "laptop", IP: "192.168.1.100", IsWired: false},
			},
			Meta: Meta{RC: "ok"},
		})
	})

	mux.HandleFunc("/proxy/network/api/s/default/rest/user", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[ClientDevice]{
			Data: []ClientDevice{
				{MAC: "CC:DD:EE:FF:00:11", Hostname: "laptop", IsWired: false},
				{MAC: "DD:EE:FF:00:11:22", Hostname: "desktop", IsWired: true},
				{MAC: "EE:FF:00:11:22:33", Hostname: "phone", IsWired: false},
			},
			Meta: Meta{RC: "ok"},
		})
	})

	mux.HandleFunc("/proxy/network/api/s/default/cmd/stamgr", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[any]{Meta: Meta{RC: "ok"}})
	})

	mux.HandleFunc("/proxy/network/api/s/default/rest/networkconf", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[Network]{
			Data: []Network{
				{ID: "net1", Name: "LAN", Purpose: "corporate", IPSubnet: "192.168.1.0/24"},
				{ID: "net2", Name: "IoT", Purpose: "corporate", IPSubnet: "192.168.2.0/24", VLAN: 20, VLANEnabled: true},
			},
			Meta: Meta{RC: "ok"},
		})
	})

	mux.HandleFunc("/proxy/network/api/s/default/rest/networkconf/net1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[Network]{
			Data: []Network{
				{ID: "net1", Name: "LAN", Purpose: "corporate", IPSubnet: "192.168.1.0/24", DHCPDEnabled: true},
			},
			Meta: Meta{RC: "ok"},
		})
	})

	mux.HandleFunc("/proxy/network/api/s/default/stat/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[HealthEntry]{
			Data: []HealthEntry{
				{Subsystem: "www", Status: "ok", NumUser: 10, Latency: 5},
				{Subsystem: "lan", Status: "ok", NumAdopted: 3},
			},
			Meta: Meta{RC: "ok"},
		})
	})

	mux.HandleFunc("/proxy/network/api/s/default/stat/sysinfo", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response[SysInfo]{
			Data: []SysInfo{
				{Name: "Default", Version: "8.0.0", Hostname: "unifi", Timezone: "UTC"},
			},
			Meta: Meta{RC: "ok"},
		})
	})

	return httptest.NewServer(mux)
}

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	apiClient, err := api.NewClient(api.Credentials{
		Host:     server.URL,
		Username: "admin",
		Password: "secret",
		Site:     "default",
	})
	if err != nil {
		t.Fatalf("create api client: %v", err)
	}

	return NewClient(apiClient)
}

func TestDevices_List(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	devices, err := client.Devices().List(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}

	if devices[0].MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected first device MAC 'AA:BB:CC:DD:EE:FF', got %q", devices[0].MAC)
	}
}

func TestDevices_List_FilterByType(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	devices, err := client.Devices().List(context.Background(), "uap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(devices) != 1 {
		t.Fatalf("expected 1 uap device, got %d", len(devices))
	}

	if devices[0].Type != "uap" {
		t.Errorf("expected type 'uap', got %q", devices[0].Type)
	}
}

func TestDevices_Get(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	device, err := client.Devices().Get(context.Background(), "AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if device.MAC != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected MAC 'AA:BB:CC:DD:EE:FF', got %q", device.MAC)
	}

	if device.NumSta != 5 {
		t.Errorf("expected 5 clients, got %d", device.NumSta)
	}
}

func TestDevices_Get_EmptyMAC(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	_, err := client.Devices().Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty MAC")
	}
}

func TestDevices_Restart(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	err := client.Devices().Restart(context.Background(), "AA:BB:CC:DD:EE:FF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDevices_Restart_EmptyMAC(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	err := client.Devices().Restart(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty MAC")
	}
}

func TestClients_ListActive(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	clients, err := client.Clients().ListActive(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(clients))
	}
}

func TestClients_ListActive_FilterWired(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	clients, err := client.Clients().ListActive(context.Background(), "wired")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(clients) != 1 {
		t.Fatalf("expected 1 wired client, got %d", len(clients))
	}

	if !clients[0].IsWired {
		t.Error("expected wired client")
	}
}

func TestClients_ListActive_FilterWireless(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	clients, err := client.Clients().ListActive(context.Background(), "wireless")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(clients) != 1 {
		t.Fatalf("expected 1 wireless client, got %d", len(clients))
	}

	if clients[0].IsWired {
		t.Error("expected wireless client")
	}
}

func TestClients_ListAll(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	clients, err := client.Clients().ListAll(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(clients) != 3 {
		t.Fatalf("expected 3 clients, got %d", len(clients))
	}
}

func TestClients_Get(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	result, err := client.Clients().Get(context.Background(), "CC:DD:EE:FF:00:11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Hostname != "laptop" {
		t.Errorf("expected hostname 'laptop', got %q", result.Hostname)
	}
}

func TestClients_Get_EmptyMAC(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	_, err := client.Clients().Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty MAC")
	}
}

func TestClients_Block(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	err := client.Clients().Block(context.Background(), "CC:DD:EE:FF:00:11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClients_Block_EmptyMAC(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	err := client.Clients().Block(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty MAC")
	}
}

func TestClients_Unblock(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	err := client.Clients().Unblock(context.Background(), "CC:DD:EE:FF:00:11")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClients_Unblock_EmptyMAC(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	err := client.Clients().Unblock(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty MAC")
	}
}

func TestNetworks_List(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	networks, err := client.Networks().List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(networks) != 2 {
		t.Fatalf("expected 2 networks, got %d", len(networks))
	}

	if networks[0].Name != "LAN" {
		t.Errorf("expected first network name 'LAN', got %q", networks[0].Name)
	}

	if networks[1].VLAN != 20 {
		t.Errorf("expected second network VLAN 20, got %d", networks[1].VLAN)
	}
}

func TestNetworks_Get(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	network, err := client.Networks().Get(context.Background(), "net1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if network.Name != "LAN" {
		t.Errorf("expected name 'LAN', got %q", network.Name)
	}

	if !network.DHCPDEnabled {
		t.Error("expected DHCP enabled")
	}
}

func TestNetworks_Get_EmptyID(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	_, err := client.Networks().Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

func TestStats_Health(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	health, err := client.Stats().Health(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(health) != 2 {
		t.Fatalf("expected 2 health entries, got %d", len(health))
	}

	if health[0].Subsystem != "www" {
		t.Errorf("expected subsystem 'www', got %q", health[0].Subsystem)
	}

	if health[0].Latency != 5 {
		t.Errorf("expected latency 5, got %d", health[0].Latency)
	}
}

func TestStats_SysInfo(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	client := newTestClient(t, server)

	sysinfo, err := client.Stats().SysInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sysinfo) != 1 {
		t.Fatalf("expected 1 sysinfo entry, got %d", len(sysinfo))
	}

	if sysinfo[0].Version != "8.0.0" {
		t.Errorf("expected version '8.0.0', got %q", sysinfo[0].Version)
	}
}
