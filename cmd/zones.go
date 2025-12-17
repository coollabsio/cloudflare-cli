package cmd

import (
	"context"

	"github.com/coollabsio/cf/internal/client"
	"github.com/coollabsio/cf/internal/output"
	"github.com/spf13/cobra"
)

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "Zone management commands",
}

var zonesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones",
	Long: `List all zones accessible by the current credentials.

Note: If your API token is scoped to specific zones, you may get a permission error.
In that case, you'll need to either:
  1. Use the zone ID directly with other commands
  2. Grant your token "All zones" read permission`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		zones, err := c.ListZones(ctx)
		if err != nil {
			return err
		}

		if len(zones) == 0 {
			out.WriteSuccess("No zones found")
			return nil
		}

		headers := []string{"ID", "Name", "Status"}
		var rows [][]string
		for _, z := range zones {
			rows = append(rows, []string{z.ID, z.Name, z.Status})
		}

		return out.WriteTable(headers, rows)
	},
}

var zonesGetCmd = &cobra.Command{
	Use:   "get <zone-name-or-id>",
	Short: "Get zone details",
	Long: `Get details for a specific zone by name or ID.

Examples:
  cf zones get example.com
  cf zones get 023e105f4ecef8ad9ca31a8372d0c353

Note: Looking up zones by name requires the "zone:list" permission.
If you have a zone-specific token, use the zone ID directly.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		zone, err := c.GetZone(ctx, args[0])
		if err != nil {
			return err
		}

		if outputFormat == "json" {
			return out.WriteJSON(zone)
		}

		headers := []string{"ID", "Name", "Status"}
		rows := [][]string{{zone.ID, zone.Name, zone.Status}}
		return out.WriteTable(headers, rows)
	},
}

func init() {
	rootCmd.AddCommand(zonesCmd)
	zonesCmd.AddCommand(zonesListCmd)
	zonesCmd.AddCommand(zonesGetCmd)
}

// resolveZone is a helper to resolve a zone argument to a zone ID
// It provides helpful error messages for permission issues
func resolveZone(c *client.Client, ctx context.Context, nameOrID string) (string, error) {
	return c.ResolveZoneID(ctx, nameOrID)
}

// mustResolveZone resolves a zone and exits on error with formatted output
func mustResolveZone(c *client.Client, ctx context.Context, nameOrID string) string {
	zoneID, err := resolveZone(c, ctx, nameOrID)
	if err != nil {
		return ""
	}
	return zoneID
}

// writeZoneTable writes zones in table format
func writeZoneTable(zones []client.Zone) error {
	headers := []string{"ID", "Name", "Status"}
	var rows [][]string
	for _, z := range zones {
		rows = append(rows, []string{z.ID, z.Name, z.Status})
	}
	return out.WriteTable(headers, rows)
}

// writeDNSRecordTable writes DNS records in table format
func writeDNSRecordTable(records []client.DNSRecord) error {
	headers := []string{"ID", "Type", "Name", "Content", "TTL", "Proxied"}
	var rows [][]string
	for _, r := range records {
		rows = append(rows, []string{
			r.ID,
			r.Type,
			r.Name,
			r.Content,
			output.FormatTTL(r.TTL),
			output.FormatBool(r.Proxied),
		})
	}
	return out.WriteTable(headers, rows)
}
