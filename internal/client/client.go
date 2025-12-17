package client

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/coollabsio/cloudflare-cli/internal/config"
)

// Client wraps the Cloudflare API client with convenience methods
type Client struct {
	api *cloudflare.API
}

// New creates a new Cloudflare client from the given config
func New(cfg *config.Config) (*Client, error) {
	if !cfg.HasCredentials() {
		return nil, errors.New("no credentials configured. Set CLOUDFLARE_API_TOKEN or CLOUDFLARE_API_KEY + CLOUDFLARE_API_EMAIL")
	}

	var api *cloudflare.API
	var err error

	if cfg.APIToken != "" {
		api, err = cloudflare.NewWithAPIToken(cfg.APIToken)
	} else {
		api, err = cloudflare.New(cfg.APIKey, cfg.APIEmail)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	return &Client{api: api}, nil
}

// VerifyToken verifies the API credentials are valid
func (c *Client) VerifyToken(ctx context.Context) error {
	// Try to verify the token
	_, err := c.api.VerifyAPIToken(ctx)
	if err != nil {
		// Fallback to listing zones if token verification fails
		_, err = c.api.ListZones(ctx)
		if err != nil {
			return fmt.Errorf("failed to verify credentials: %w", err)
		}
	}
	return nil
}

// Zone represents a Cloudflare zone
type Zone struct {
	ID     string
	Name   string
	Status string
}

// ListZones returns all zones accessible by the current credentials
func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	zones, err := c.api.ListZones(ctx)
	if err != nil {
		if isPermissionError(err) {
			return nil, fmt.Errorf("permission denied: your API token may not have 'Zone:Read' permission for all zones. %w", err)
		}
		return nil, err
	}

	var result []Zone
	for _, z := range zones {
		result = append(result, Zone{
			ID:     z.ID,
			Name:   z.Name,
			Status: z.Status,
		})
	}
	return result, nil
}

// GetZone returns a zone by name or ID
func (c *Client) GetZone(ctx context.Context, nameOrID string) (*Zone, error) {
	// First, try to get by ID directly (works with zone-specific tokens)
	if looksLikeZoneID(nameOrID) {
		zone, err := c.api.ZoneDetails(ctx, nameOrID)
		if err == nil {
			return &Zone{
				ID:     zone.ID,
				Name:   zone.Name,
				Status: zone.Status,
			}, nil
		}
		// If it failed, it might not be an ID after all, try by name
	}

	// Try to find by name
	zones, err := c.api.ListZones(ctx, nameOrID)
	if err != nil {
		if isPermissionError(err) {
			return nil, fmt.Errorf(`permission denied when looking up zone by name.

This usually happens when your API token is scoped to specific zones.
To fix this, either:
  1. Use the zone ID directly: cf zones get <zone-id>
  2. Grant your token "All zones" read permission

Error: %w`, err)
		}
		return nil, err
	}

	if len(zones) == 0 {
		return nil, fmt.Errorf("zone not found: %s", nameOrID)
	}

	z := zones[0]
	return &Zone{
		ID:     z.ID,
		Name:   z.Name,
		Status: z.Status,
	}, nil
}

// ResolveZoneID resolves a zone name or ID to a zone ID
func (c *Client) ResolveZoneID(ctx context.Context, nameOrID string) (string, error) {
	zone, err := c.GetZone(ctx, nameOrID)
	if err != nil {
		return "", err
	}
	return zone.ID, nil
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	ID       string
	Type     string
	Name     string
	Content  string
	TTL      int
	Proxied  bool
	Priority *uint16
	Comment  string
}

// ListDNSRecords returns DNS records for a zone
func (c *Client) ListDNSRecords(ctx context.Context, zoneID string, recordType, name string) ([]DNSRecord, error) {
	filter := cloudflare.DNSRecord{}
	if recordType != "" {
		filter.Type = recordType
	}
	if name != "" {
		filter.Name = name
	}

	rc := cloudflare.ZoneIdentifier(zoneID)
	records, _, err := c.api.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{
		Type: filter.Type,
		Name: filter.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}

	var result []DNSRecord
	for _, r := range records {
		rec := DNSRecord{
			ID:       r.ID,
			Type:     r.Type,
			Name:     r.Name,
			Content:  r.Content,
			TTL:      r.TTL,
			Proxied:  boolValue(r.Proxied),
			Priority: r.Priority,
			Comment:  r.Comment,
		}
		result = append(result, rec)
	}
	return result, nil
}

// GetDNSRecord returns a specific DNS record
func (c *Client) GetDNSRecord(ctx context.Context, zoneID, recordID string) (*DNSRecord, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	r, err := c.api.GetDNSRecord(ctx, rc, recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record: %w", err)
	}

	return &DNSRecord{
		ID:       r.ID,
		Type:     r.Type,
		Name:     r.Name,
		Content:  r.Content,
		TTL:      r.TTL,
		Proxied:  boolValue(r.Proxied),
		Priority: r.Priority,
		Comment:  r.Comment,
	}, nil
}

// CreateDNSRecordParams contains parameters for creating a DNS record
type CreateDNSRecordParams struct {
	Type     string
	Name     string
	Content  string
	TTL      int
	Proxied  bool
	Priority *uint16
	Comment  string
}

// CreateDNSRecord creates a new DNS record
func (c *Client) CreateDNSRecord(ctx context.Context, zoneID string, params CreateDNSRecordParams) (*DNSRecord, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)

	createParams := cloudflare.CreateDNSRecordParams{
		Type:     params.Type,
		Name:     params.Name,
		Content:  params.Content,
		TTL:      params.TTL,
		Proxied:  &params.Proxied,
		Priority: params.Priority,
		Comment:  params.Comment,
	}

	r, err := c.api.CreateDNSRecord(ctx, rc, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}

	return &DNSRecord{
		ID:       r.ID,
		Type:     r.Type,
		Name:     r.Name,
		Content:  r.Content,
		TTL:      r.TTL,
		Proxied:  boolValue(r.Proxied),
		Priority: r.Priority,
		Comment:  r.Comment,
	}, nil
}

// UpdateDNSRecordParams contains parameters for updating a DNS record
type UpdateDNSRecordParams struct {
	Type     string
	Name     string
	Content  string
	TTL      *int
	Proxied  *bool
	Priority *uint16
	Comment  *string
}

// UpdateDNSRecord updates an existing DNS record
func (c *Client) UpdateDNSRecord(ctx context.Context, zoneID, recordID string, params UpdateDNSRecordParams) (*DNSRecord, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)

	updateParams := cloudflare.UpdateDNSRecordParams{
		ID:       recordID,
		Type:     params.Type,
		Name:     params.Name,
		Content:  params.Content,
		Proxied:  params.Proxied,
		Priority: params.Priority,
		Comment:  params.Comment,
	}

	if params.TTL != nil {
		updateParams.TTL = *params.TTL
	}

	r, err := c.api.UpdateDNSRecord(ctx, rc, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update DNS record: %w", err)
	}

	return &DNSRecord{
		ID:       r.ID,
		Type:     r.Type,
		Name:     r.Name,
		Content:  r.Content,
		TTL:      r.TTL,
		Proxied:  boolValue(r.Proxied),
		Priority: r.Priority,
		Comment:  r.Comment,
	}, nil
}

// DeleteDNSRecord deletes a DNS record
func (c *Client) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	rc := cloudflare.ZoneIdentifier(zoneID)
	err := c.api.DeleteDNSRecord(ctx, rc, recordID)
	if err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}
	return nil
}

// FindDNSRecords finds DNS records by name and type
func (c *Client) FindDNSRecords(ctx context.Context, zoneID, name, recordType string) ([]DNSRecord, error) {
	return c.ListDNSRecords(ctx, zoneID, recordType, name)
}

// boolValue safely dereferences a bool pointer
func boolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// looksLikeZoneID checks if the string looks like a Cloudflare zone ID (32 hex chars)
func looksLikeZoneID(s string) bool {
	if len(s) != 32 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// isPermissionError checks if the error is a permission/authorization error
func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "403")
}
