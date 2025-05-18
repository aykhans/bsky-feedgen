package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aykhans/bsky-feedgen/pkg/manage"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go listenForTermination(func() {
		cancel()
		fmt.Println()
		os.Exit(130)
	})

	var rootCmd = &cobra.Command{
		Use:               "feedgen-manager",
		Short:             "BlueSkye Feed Generator client CLI",
		Long:              "A command-line tool for managing feed generators on Bluesky",
		CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	}

	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a feed generator record",
		Run: func(cmd *cobra.Command, args []string) {
			handle := prompt("Enter your Bluesky handle", "")
			if handle == "" {
				fmt.Println("\nError: handle is required")
				os.Exit(1)
			}

			password := promptPassword("Enter your Bluesky password (preferably an App Password)")
			if handle == "" {
				fmt.Println("\nError: password is required")
				os.Exit(1)
			}

			service := prompt("Optionally, enter a custom PDS service to sign in with", manage.DefaultPDSHost)

			feedgenHostname := prompt("Enter the feed generator hostname (e.g. 'feeds.bsky.example.com')", "")
			if feedgenHostname == "" {
				fmt.Println("\nError: hostname is required")
				os.Exit(1)
			}

			recordKey := prompt("Enter a short name for the record. This will be shown in the feed's URL and should be unique.", "")
			if recordKey == "" {
				fmt.Println("\nError: short name is required")
				os.Exit(1)
			}

			displayName := prompt("Enter a display name for your feed", "")
			if displayName == "" {
				fmt.Println("\nError: display name is required")
				os.Exit(1)
			}

			description := prompt("Optionally, enter a brief description of your feed", "")
			avatar := prompt("Optionally, enter a local path to an avatar that will be used for the feed", "")

			client, err := manage.NewClientWithAuth(ctx, manage.NewClient(&service), handle, password)
			if err != nil {
				fmt.Printf("\nAuthentication failed: %v", err)
				os.Exit(1)
			}

			err = manage.CreateFeedGenerator(
				ctx,
				client,
				displayName,
				toPtr(description),
				toPtr(avatar),
				"did:web:"+feedgenHostname,
				recordKey,
			)
			if err != nil {
				fmt.Printf("\nFailed to create feed generator record: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("\nFeed generator created successfully! 🎉")
		},
	}

	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update a feed generator record",
		Run: func(cmd *cobra.Command, args []string) {
			handle := prompt("Enter your Bluesky handle", "")
			if handle == "" {
				fmt.Println("\nError: handle is required")
				os.Exit(1)
			}

			password := promptPassword("Enter your Bluesky password (preferably an App Password)")
			if handle == "" {
				fmt.Println("\nError: password is required")
				os.Exit(1)
			}

			service := prompt("Optionally, enter a custom PDS service to sign in with", manage.DefaultPDSHost)
			feedgenHostname := prompt("Optionally, enter the feed generator hostname (e.g. 'feeds.bsky.example.com')", "")

			recordKey := prompt("Enter short name of the record", "")
			if recordKey == "" {
				fmt.Println("\nError: short name is required")
				os.Exit(1)
			}

			displayName := prompt("Optionally, enter a display name for your feed", "")
			description := prompt("Optionally, enter a brief description of your feed", "")
			avatar := prompt("Optionally, enter a local path to an avatar that will be used for the feed", "")

			client, err := manage.NewClientWithAuth(ctx, manage.NewClient(&service), handle, password)
			if err != nil {
				fmt.Printf("\nAuthentication failed: %v", err)
				os.Exit(1)
			}

			var did *string
			if feedgenHostname != "" {
				did = toPtr("did:web:" + feedgenHostname)
			}

			err = manage.UpdateFeedGenerator(
				ctx,
				client,
				toPtr(displayName),
				toPtr(description),
				toPtr(avatar),
				did,
				recordKey,
			)
			if err != nil {
				fmt.Printf("\nFailed to update feed generator record: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("\nFeed generator updated successfully! 🎉")
		},
	}

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a feed generator record",
		Run: func(cmd *cobra.Command, args []string) {
			handle := prompt("Enter your Bluesky handle", "")
			if handle == "" {
				fmt.Println("\nError: handle is required")
				os.Exit(1)
			}

			password := promptPassword("Enter your Bluesky password (preferably an App Password)")
			if handle == "" {
				fmt.Println("\nError: password is required")
				os.Exit(1)
			}

			service := prompt("Optionally, enter a custom PDS service to sign in with", manage.DefaultPDSHost)

			recordKey := prompt("Enter short name of the record", "")
			if recordKey == "" {
				fmt.Println("\nError: short name is required")
				os.Exit(1)
			}

			confirm := promptConfirm("Are you sure you want to delete this record? Any likes that your feed has will be lost", false)
			if !confirm {
				fmt.Println("\nAborting...")
				return
			}

			client, err := manage.NewClientWithAuth(ctx, manage.NewClient(&service), handle, password)
			if err != nil {
				fmt.Printf("\nAuthentication failed: %v", err)
				os.Exit(1)
			}

			err = manage.DeleteFeedGenerator(ctx, client, recordKey)
			if err != nil {
				fmt.Printf("\nFailed to delete feed generator record: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("\nFeed generator deleted successfully! 🎉")
		},
	}

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// <-------------- Utils -------------->

func listenForTermination(do func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	do()
}

func toPtr[T comparable](value T) *T {
	var zero T
	if value == zero {
		return nil
	}

	return &value
}

func prompt(label string, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", label, defaultValue)
	} else {
		fmt.Printf("%s: ", label)
	}

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultValue
	}
	return input
}

func promptPassword(label string) string {
	fmt.Printf("%s: ", label)
	password, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password)
}

func promptConfirm(label string, defaultValue bool) bool {
	defaultStr := "y/N"
	if defaultValue {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s [%s]: ", label, defaultStr)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" {
		return defaultValue
	}

	return input == "y" || input == "yes"
}
