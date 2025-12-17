package cmd

import (
	"context"
	"fmt"

	"github.com/heyandras/cfdns/internal/client"
	"github.com/heyandras/cfdns/internal/output"
	"github.com/spf13/cobra"
)

var (
	dnsType     string
	dnsName     string
	dnsContent  string
	dnsTTL      int
	dnsProxied  string
	dnsPriority uint16
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS record management commands",
}

var dnsListCmd = &cobra.Command{
	Use:   "list <zone>",
	Short: "List DNS records",
	Long: `List DNS records for a zone.

Examples:
  cfdns dns list example.com
  cfdns dns list example.com --type A
  cfdns dns list example.com --name www
  cfdns dns list 023e105f4ecef8ad9ca31a8372d0c353`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		zoneID, err := resolveZone(c, ctx, args[0])
		if err != nil {
			return err
		}

		records, err := c.ListDNSRecords(ctx, zoneID, dnsType, dnsName)
		if err != nil {
			return err
		}

		if len(records) == 0 {
			out.WriteSuccess("No DNS records found")
			return nil
		}

		return writeDNSRecordTable(records)
	},
}

var dnsGetCmd = &cobra.Command{
	Use:   "get <zone> <record-id>",
	Short: "Get DNS record details",
	Long: `Get details for a specific DNS record.

Example:
  cfdns dns get example.com 372e67954025e0ba6aaa6d586b9e0b59`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		zoneID, err := resolveZone(c, ctx, args[0])
		if err != nil {
			return err
		}

		record, err := c.GetDNSRecord(ctx, zoneID, args[1])
		if err != nil {
			return err
		}

		if outputFormat == "json" {
			return out.WriteJSON(record)
		}

		headers := []string{"ID", "Type", "Name", "Content", "TTL", "Proxied"}
		rows := [][]string{{
			record.ID,
			record.Type,
			record.Name,
			record.Content,
			output.FormatTTL(record.TTL),
			output.FormatBool(record.Proxied),
		}}
		return out.WriteTable(headers, rows)
	},
}

var dnsCreateCmd = &cobra.Command{
	Use:   "create <zone>",
	Short: "Create a DNS record",
	Long: `Create a new DNS record.

Examples:
  cfdns dns create example.com --name www --type A --content 192.0.2.1
  cfdns dns create example.com --name www --type CNAME --content example.com --proxied
  cfdns dns create example.com --name mail --type MX --content mail.example.com --priority 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dnsType == "" || dnsName == "" || dnsContent == "" {
			return fmt.Errorf("--type, --name, and --content are required")
		}

		// Parse proxied flag
		proxied := false
		if dnsProxied != "" {
			if dnsProxied != "true" && dnsProxied != "false" {
				return fmt.Errorf("--proxied must be 'true' or 'false'")
			}
			proxied = dnsProxied == "true"
		}

		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		zoneID, err := resolveZone(c, ctx, args[0])
		if err != nil {
			return err
		}

		params := client.CreateDNSRecordParams{
			Type:    dnsType,
			Name:    dnsName,
			Content: dnsContent,
			TTL:     dnsTTL,
			Proxied: proxied,
		}
		if dnsPriority > 0 {
			params.Priority = &dnsPriority
		}

		record, err := c.CreateDNSRecord(ctx, zoneID, params)
		if err != nil {
			return err
		}

		if outputFormat == "json" {
			return out.WriteJSON(record)
		}

		out.WriteSuccess(fmt.Sprintf("Created DNS record: %s", record.ID))
		headers := []string{"ID", "Type", "Name", "Content", "TTL", "Proxied"}
		rows := [][]string{{
			record.ID,
			record.Type,
			record.Name,
			record.Content,
			output.FormatTTL(record.TTL),
			output.FormatBool(record.Proxied),
		}}
		return out.WriteTable(headers, rows)
	},
}

var dnsUpdateCmd = &cobra.Command{
	Use:   "update <zone> <record-id>",
	Short: "Update a DNS record",
	Long: `Update an existing DNS record.

Only specify the fields you want to change. Unspecified fields keep their current values.

Examples:
  cf dns update example.com 372e67954025e0ba6aaa6d586b9e0b59 --content 192.0.2.2
  cf dns update example.com 372e67954025e0ba6aaa6d586b9e0b59 --name www2
  cf dns update example.com 372e67954025e0ba6aaa6d586b9e0b59 --proxied
  cf dns update example.com 372e67954025e0ba6aaa6d586b9e0b59 --proxied=false`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		zoneID, err := resolveZone(c, ctx, args[0])
		if err != nil {
			return err
		}

		// Fetch existing record first
		existing, err := c.GetDNSRecord(ctx, zoneID, args[1])
		if err != nil {
			return err
		}

		// Start with existing values
		params := client.UpdateDNSRecordParams{
			Type:    existing.Type,
			Name:    existing.Name,
			Content: existing.Content,
		}

		// Override only the fields that were explicitly set
		if cmd.Flags().Changed("type") {
			params.Type = dnsType
		}
		if cmd.Flags().Changed("name") {
			params.Name = dnsName
		}
		if cmd.Flags().Changed("content") {
			params.Content = dnsContent
		}
		if cmd.Flags().Changed("ttl") {
			params.TTL = &dnsTTL
		}
		if cmd.Flags().Changed("proxied") {
			if dnsProxied != "true" && dnsProxied != "false" {
				return fmt.Errorf("--proxied must be 'true' or 'false'")
			}
			proxied := dnsProxied == "true"
			params.Proxied = &proxied
		}
		if cmd.Flags().Changed("priority") {
			params.Priority = &dnsPriority
		}

		record, err := c.UpdateDNSRecord(ctx, zoneID, args[1], params)
		if err != nil {
			return err
		}

		if outputFormat == "json" {
			return out.WriteJSON(record)
		}

		out.WriteSuccess(fmt.Sprintf("Updated DNS record: %s", record.ID))
		headers := []string{"ID", "Type", "Name", "Content", "TTL", "Proxied"}
		rows := [][]string{{
			record.ID,
			record.Type,
			record.Name,
			record.Content,
			output.FormatTTL(record.TTL),
			output.FormatBool(record.Proxied),
		}}
		return out.WriteTable(headers, rows)
	},
}

var dnsDeleteCmd = &cobra.Command{
	Use:   "delete <zone> <record-id>",
	Short: "Delete a DNS record",
	Long: `Delete a DNS record.

Example:
  cfdns dns delete example.com 372e67954025e0ba6aaa6d586b9e0b59`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		zoneID, err := resolveZone(c, ctx, args[0])
		if err != nil {
			return err
		}

		if err := c.DeleteDNSRecord(ctx, zoneID, args[1]); err != nil {
			return err
		}

		out.WriteSuccess(fmt.Sprintf("Deleted DNS record: %s", args[1]))
		return nil
	},
}

var dnsFindCmd = &cobra.Command{
	Use:   "find <zone>",
	Short: "Find DNS records by name and type",
	Long: `Find DNS records by name and/or type. Useful for getting record IDs.

Examples:
  cfdns dns find example.com --name www --type A
  cfdns dns find example.com --name mail --type MX`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if dnsName == "" && dnsType == "" {
			return fmt.Errorf("at least one of --name or --type is required")
		}

		c, err := client.New(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		zoneID, err := resolveZone(c, ctx, args[0])
		if err != nil {
			return err
		}

		records, err := c.FindDNSRecords(ctx, zoneID, dnsName, dnsType)
		if err != nil {
			return err
		}

		if len(records) == 0 {
			out.WriteSuccess("No matching DNS records found")
			return nil
		}

		return writeDNSRecordTable(records)
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	// List command
	dnsListCmd.Flags().StringVarP(&dnsType, "type", "t", "", "filter by record type (A, AAAA, CNAME, TXT, MX, etc.)")
	dnsListCmd.Flags().StringVarP(&dnsName, "name", "n", "", "filter by record name")
	dnsCmd.AddCommand(dnsListCmd)

	// Get command
	dnsCmd.AddCommand(dnsGetCmd)

	// Create command
	dnsCreateCmd.Flags().StringVarP(&dnsType, "type", "t", "", "record type (required)")
	dnsCreateCmd.Flags().StringVarP(&dnsName, "name", "n", "", "record name (required)")
	dnsCreateCmd.Flags().StringVarP(&dnsContent, "content", "c", "", "record content (required)")
	dnsCreateCmd.Flags().IntVar(&dnsTTL, "ttl", 1, "TTL in seconds (1 = auto)")
	dnsCreateCmd.Flags().StringVar(&dnsProxied, "proxied", "", "proxy through Cloudflare (true|false)")
	dnsCreateCmd.Flags().Lookup("proxied").NoOptDefVal = "true"
	dnsCreateCmd.Flags().Uint16Var(&dnsPriority, "priority", 0, "record priority (for MX, SRV)")
	dnsCmd.AddCommand(dnsCreateCmd)

	// Update command
	dnsUpdateCmd.Flags().StringVarP(&dnsType, "type", "t", "", "new record type")
	dnsUpdateCmd.Flags().StringVarP(&dnsName, "name", "n", "", "new record name")
	dnsUpdateCmd.Flags().StringVarP(&dnsContent, "content", "c", "", "new record content")
	dnsUpdateCmd.Flags().IntVar(&dnsTTL, "ttl", 1, "TTL in seconds (1 = auto)")
	dnsUpdateCmd.Flags().StringVar(&dnsProxied, "proxied", "", "set proxy status (true|false)")
	dnsUpdateCmd.Flags().Lookup("proxied").NoOptDefVal = "true"
	dnsUpdateCmd.Flags().Uint16Var(&dnsPriority, "priority", 0, "record priority (for MX, SRV)")
	dnsCmd.AddCommand(dnsUpdateCmd)

	// Delete command
	dnsCmd.AddCommand(dnsDeleteCmd)

	// Find command
	dnsFindCmd.Flags().StringVarP(&dnsType, "type", "t", "", "record type to find")
	dnsFindCmd.Flags().StringVarP(&dnsName, "name", "n", "", "record name to find")
	dnsCmd.AddCommand(dnsFindCmd)
}
