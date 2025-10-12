package cli

import (
	"fmt"
	"strings"

	"github.com/binance-live/internal/database"
	"github.com/binance-live/internal/models"
	"github.com/binance-live/internal/repository"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewSymbolsCmd() *cobra.Command {
	symbolsCmd := &cobra.Command{
		Use:   "symbols",
		Short: "Symbol management commands",
		Long:  `Commands for managing trading symbols`,
	}

	symbolsCmd.AddCommand(NewListSymbolsCmd())
	symbolsCmd.AddCommand(NewAddSymbolCmd())
	symbolsCmd.AddCommand(NewDeactivateSymbolCmd())
	symbolsCmd.AddCommand(NewActivateSymbolCmd())

	return symbolsCmd
}

func NewListSymbolsCmd() *cobra.Command {
	var activeOnly bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List symbols",
		Long:  `List all symbols or only active symbols`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListSymbols(activeOnly)
		},
	}

	cmd.Flags().BoolVarP(&activeOnly, "active-only", "a", false, "Show only active symbols")

	return cmd
}

func NewAddSymbolCmd() *cobra.Command {
	var (
		symbol     string
		baseAsset  string
		quoteAsset string
		status     string
		isActive   bool
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new symbol",
		Long:  `Add a new trading symbol to the database`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if symbol == "" || baseAsset == "" || quoteAsset == "" {
				return fmt.Errorf("symbol, base-asset, and quote-asset are required")
			}
			return runAddSymbol(symbol, baseAsset, quoteAsset, status, isActive)
		},
	}

	cmd.Flags().StringVarP(&symbol, "symbol", "s", "", "Symbol name (required)")
	cmd.Flags().StringVarP(&baseAsset, "base-asset", "b", "", "Base asset (required)")
	cmd.Flags().StringVarP(&quoteAsset, "quote-asset", "q", "", "Quote asset (required)")
	cmd.Flags().StringVar(&status, "status", "TRADING", "Symbol status")
	cmd.Flags().BoolVar(&isActive, "active", true, "Set symbol as active")

	cmd.MarkFlagRequired("symbol")
	cmd.MarkFlagRequired("base-asset")
	cmd.MarkFlagRequired("quote-asset")

	return cmd
}

func NewDeactivateSymbolCmd() *cobra.Command {
	var symbol string

	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivate a symbol",
		Long:  `Deactivate a trading symbol to stop data collection`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if symbol == "" {
				return fmt.Errorf("symbol is required")
			}
			return runUpdateSymbolStatus(symbol, false)
		},
	}

	cmd.Flags().StringVarP(&symbol, "symbol", "s", "", "Symbol to deactivate (required)")
	cmd.MarkFlagRequired("symbol")

	return cmd
}

func NewActivateSymbolCmd() *cobra.Command {
	var symbol string

	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate a symbol",
		Long:  `Activate a trading symbol to start data collection`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if symbol == "" {
				return fmt.Errorf("symbol is required")
			}
			return runUpdateSymbolStatus(symbol, true)
		},
	}

	cmd.Flags().StringVarP(&symbol, "symbol", "s", "", "Symbol to activate (required)")
	cmd.MarkFlagRequired("symbol")

	return cmd
}

func runListSymbols(activeOnly bool) error {
	cfg, log, ctx, err := getSharedResources()
	if err != nil {
		return err
	}
	defer log.Sync()

	// Initialize database
	db, err := database.New(&cfg.Database, log)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repository
	symbolRepo := repository.NewSymbolRepository(db)

	var symbols []models.Symbol
	if activeOnly {
		symbols, err = symbolRepo.GetActiveSymbols(ctx)
		if err != nil {
			return fmt.Errorf("failed to get active symbols: %w", err)
		}
	} else {
		// Note: You'd need to add a GetAllSymbols method to the repository
		// For now, we'll just get active symbols
		symbols, err = symbolRepo.GetActiveSymbols(ctx)
		if err != nil {
			return fmt.Errorf("failed to get symbols: %w", err)
		}
	}

	// Print symbols
	fmt.Printf("Found %d symbols:\n\n", len(symbols))
	fmt.Printf("%-15s %-8s %-8s %-10s %-8s\n", "SYMBOL", "BASE", "QUOTE", "STATUS", "ACTIVE")
	fmt.Println(strings.Repeat("-", 55))

	for _, sym := range symbols {
		activeStatus := "NO"
		if sym.IsActive {
			activeStatus = "YES"
		}
		fmt.Printf("%-15s %-8s %-8s %-10s %-8s\n",
			sym.Symbol, sym.BaseAsset, sym.QuoteAsset, sym.Status, activeStatus)
	}

	return nil
}

func runAddSymbol(symbol, baseAsset, quoteAsset, status string, isActive bool) error {
	cfg, log, ctx, err := getSharedResources()
	if err != nil {
		return err
	}
	defer log.Sync()

	// Normalize inputs
	symbol = strings.ToUpper(symbol)
	baseAsset = strings.ToUpper(baseAsset)
	quoteAsset = strings.ToUpper(quoteAsset)
	status = strings.ToUpper(status)

	log.Info("Adding symbol",
		zap.String("symbol", symbol),
		zap.String("base_asset", baseAsset),
		zap.String("quote_asset", quoteAsset),
		zap.String("status", status),
		zap.Bool("is_active", isActive),
	)

	// Initialize database
	db, err := database.New(&cfg.Database, log)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repository
	symbolRepo := repository.NewSymbolRepository(db)

	// Create symbol
	newSymbol := &models.Symbol{
		Symbol:     symbol,
		BaseAsset:  baseAsset,
		QuoteAsset: quoteAsset,
		Status:     status,
		IsActive:   isActive,
	}

	if err := symbolRepo.UpsertSymbol(ctx, newSymbol); err != nil {
		return fmt.Errorf("failed to add symbol: %w", err)
	}

	fmt.Printf("Successfully added symbol %s (Base: %s, Quote: %s, Active: %v)\n",
		symbol, baseAsset, quoteAsset, isActive)

	return nil
}

func runUpdateSymbolStatus(symbol string, isActive bool) error {
	cfg, log, ctx, err := getSharedResources()
	if err != nil {
		return err
	}
	defer log.Sync()

	symbol = strings.ToUpper(symbol)

	action := "Deactivating"
	if isActive {
		action = "Activating"
	}

	log.Info("Updating symbol status",
		zap.String("symbol", symbol),
		zap.Bool("is_active", isActive),
	)

	// Initialize database
	db, err := database.New(&cfg.Database, log)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repository
	symbolRepo := repository.NewSymbolRepository(db)

	// Check if symbol exists
	_, err = symbolRepo.GetSymbolByName(ctx, symbol)
	if err != nil {
		return fmt.Errorf("symbol %s not found: %w", symbol, err)
	}

	// Update status
	if err := symbolRepo.UpdateSymbolStatus(ctx, symbol, isActive); err != nil {
		return fmt.Errorf("failed to update symbol status: %w", err)
	}

	fmt.Printf("%s symbol %s successfully\n", action, symbol)

	return nil
}
