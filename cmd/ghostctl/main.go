package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/iammahbubalam/ghost-agent/pkg/agentpb"
)

var (
	agentAddr string
	timeout   time.Duration
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ghostctl",
		Short: "Ghost Agent CLI - Manage VMs and agent",
		Long:  `ghostctl is a command-line tool to manage Ghost Agent and VMs`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&agentAddr, "agent", "localhost:9090", "Ghost Agent gRPC address")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Second, "Request timeout")

	// Add commands
	rootCmd.AddCommand(vmCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// vmCmd returns the VM management command
func vmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vm",
		Short: "Manage virtual machines",
	}

	cmd.AddCommand(vmListCmd())
	cmd.AddCommand(vmCreateCmd())
	cmd.AddCommand(vmDeleteCmd())
	cmd.AddCommand(vmStartCmd())
	cmd.AddCommand(vmStopCmd())
	cmd.AddCommand(vmStatusCmd())

	return cmd
}

// vmListCmd lists all VMs
func vmListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all VMs",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := connectToAgent()
			if err != nil {
				return err
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			resp, err := client.ListVMs(ctx, &agentpb.ListVMsRequest{})
			if err != nil {
				return fmt.Errorf("failed to list VMs: %w", err)
			}

			if len(resp.Vms) == 0 {
				fmt.Println("No VMs found")
				return nil
			}

			fmt.Printf("%-20s %-30s %-15s %-15s\n", "VM ID", "Name", "Status", "IP Address")
			fmt.Println("--------------------------------------------------------------------------------")
			for _, vm := range resp.Vms {
				fmt.Printf("%-20s %-30s %-15s %-15s\n",
					vm.VmId, vm.Name, vm.Status, vm.IpAddress)
			}

			return nil
		},
	}
}

// vmCreateCmd creates a new VM
func vmCreateCmd() *cobra.Command {
	var (
		name     string
		vcpu     int32
		ramGB    int32
		diskGB   int32
		template string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new VM",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := connectToAgent()
			if err != nil {
				return err
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			req := &agentpb.CreateVMRequest{
				Name:     name,
				Vcpu:     vcpu,
				RamGb:    ramGB,
				DiskGb:   diskGB,
				Template: template,
			}

			fmt.Printf("Creating VM '%s'...\n", name)
			resp, err := client.CreateVM(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create VM: %w", err)
			}

			fmt.Printf("✅ VM created successfully!\n")
			fmt.Printf("  VM ID: %s\n", resp.VmId)
			fmt.Printf("  IP Address: %s\n", resp.IpAddress)
			fmt.Printf("  Status: %s\n", resp.Status)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "VM name (required)")
	cmd.Flags().Int32Var(&vcpu, "vcpu", 2, "Number of vCPUs")
	cmd.Flags().Int32Var(&ramGB, "ram", 4, "RAM in GB")
	cmd.Flags().Int32Var(&diskGB, "disk", 50, "Disk size in GB")
	cmd.Flags().StringVar(&template, "template", "ubuntu-22.04", "OS template (ubuntu-22.04, ubuntu-20.04, debian-12, debian-11)")
	cmd.MarkFlagRequired("name")

	return cmd
}

// vmDeleteCmd deletes a VM
func vmDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <vm-id>",
		Short: "Delete a VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := connectToAgent()
			if err != nil {
				return err
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			vmID := args[0]
			fmt.Printf("Deleting VM '%s'...\n", vmID)

			resp, err := client.DeleteVM(ctx, &agentpb.DeleteVMRequest{VmId: vmID})
			if err != nil {
				return fmt.Errorf("failed to delete VM: %w", err)
			}

			if resp.Success {
				fmt.Printf("✅ VM deleted successfully\n")
			} else {
				fmt.Printf("❌ Failed to delete VM\n")
			}

			return nil
		},
	}
}

// vmStartCmd starts a VM
func vmStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <vm-id>",
		Short: "Start a VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := connectToAgent()
			if err != nil {
				return err
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			vmID := args[0]
			fmt.Printf("Starting VM '%s'...\n", vmID)

			resp, err := client.StartVM(ctx, &agentpb.StartVMRequest{VmId: vmID})
			if err != nil {
				return fmt.Errorf("failed to start VM: %w", err)
			}

			fmt.Printf("✅ VM started: %s\n", resp.Status)
			return nil
		},
	}
}

// vmStopCmd stops a VM
func vmStopCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "stop <vm-id>",
		Short: "Stop a VM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := connectToAgent()
			if err != nil {
				return err
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			vmID := args[0]
			fmt.Printf("Stopping VM '%s'...\n", vmID)

			resp, err := client.StopVM(ctx, &agentpb.StopVMRequest{
				VmId:  vmID,
				Force: force,
			})
			if err != nil {
				return fmt.Errorf("failed to stop VM: %w", err)
			}

			fmt.Printf("✅ VM stopped: %s\n", resp.Status)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force stop (power off)")
	return cmd
}

// vmStatusCmd gets VM status
func vmStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <vm-id>",
		Short: "Get VM status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, conn, err := connectToAgent()
			if err != nil {
				return err
			}
			defer conn.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			vmID := args[0]
			resp, err := client.GetVMStatus(ctx, &agentpb.GetVMStatusRequest{VmId: vmID})
			if err != nil {
				return fmt.Errorf("failed to get VM status: %w", err)
			}

			fmt.Printf("VM Status:\n")
			fmt.Printf("  ID: %s\n", resp.VmId)
			fmt.Printf("  Name: %s\n", resp.Name)
			fmt.Printf("  Status: %s\n", resp.Status)
			fmt.Printf("  vCPU: %d\n", resp.Vcpu)
			fmt.Printf("  RAM: %d GB\n", resp.RamGb)
			fmt.Printf("  Disk: %d GB\n", resp.DiskGb)
			fmt.Printf("  IP Address: %s\n", resp.IpAddress)
			fmt.Printf("  Uptime: %d seconds\n", resp.UptimeSeconds)
			fmt.Printf("  CPU Usage: %.2f%%\n", resp.CpuUsagePercent)
			fmt.Printf("  RAM Usage: %.2f%%\n", resp.RamUsagePercent)

			return nil
		},
	}
}

// statusCmd shows agent status
func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Ghost Agent status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Try to connect
			_, conn, err := connectToAgent()
			if err != nil {
				fmt.Printf("❌ Ghost Agent is not running or not reachable\n")
				fmt.Printf("   Error: %v\n", err)
				return nil
			}
			defer conn.Close()

			fmt.Printf("✅ Ghost Agent is running\n")
			fmt.Printf("   Address: %s\n", agentAddr)

			return nil
		},
	}
}

// versionCmd shows version
func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show ghostctl version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ghostctl version 1.0.0")
		},
	}
}

// connectToAgent creates a gRPC connection to Ghost Agent
func connectToAgent() (agentpb.AgentServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(agentAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Ghost Agent at %s: %w", agentAddr, err)
	}

	client := agentpb.NewAgentServiceClient(conn)
	return client, conn, nil
}
