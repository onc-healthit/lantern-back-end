package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/config"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/endpointmanager/postgresql"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/xuri/excelize/v2"
)

func main() {
	var csvFilePath string

	if len(os.Args) >= 2 {
		csvFilePath = os.Args[1]
	} else {
		log.Fatalf("ERROR: Missing command-line arguments. Usage: chplsharedlistsources <csv_file_path>")
	}

	// Setup configuration using viper
	err := config.SetupConfig()
	helpers.FailOnError("Failed to setup config", err)

	// Connect to database
	store, err := postgresql.NewStore(
		viper.GetString("dbhost"),
		viper.GetInt("dbport"),
		viper.GetString("dbuser"),
		viper.GetString("dbpassword"),
		viper.GetString("dbname"),
		viper.GetString("dbsslmode"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	log.Info("Successfully connected to DB!")

	ctx := context.Background()

	// Parse CSV file
	entries, err := parseCSV(csvFilePath)
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}

	log.Infof("Parsed %d entries from CSV", len(entries))

	// Insert entries into database
	err = populateSharedListSources(ctx, store, entries)
	if err != nil {
		log.Fatalf("Failed to populate shared_list_sources table: %v", err)
	}

	log.Info("Successfully populated shared_list_sources table")
}

type SharedListSourceEntry struct {
	DeveloperName string
	ListSource    string
}

func parseCSV(csvFilePath string) ([]SharedListSourceEntry, error) {
	// Check file extension to determine if it's Excel or CSV
	ext := strings.ToLower(filepath.Ext(csvFilePath))
	log.Infof("File path: %s", csvFilePath)
	log.Infof("File extension: %s", ext)

	// Get file info for debugging
	fileInfo, err := os.Stat(csvFilePath)
	if err != nil {
		log.Errorf("Error getting file info: %v", err)
	} else {
		log.Infof("File size: %d bytes", fileInfo.Size())
	}

	// Try to detect if it's actually an Excel file by checking magic bytes
	isExcel, err := isExcelFile(csvFilePath)
	if err != nil {
		log.Errorf("Error checking file type: %v", err)
		return nil, fmt.Errorf("failed to check file type: %w", err)
	}
	log.Infof("isExcel detection result: %v", isExcel)

	// Log the actual magic bytes for debugging
	file, err := os.Open(csvFilePath)
	if err != nil {
		log.Errorf("Error opening file to read magic bytes: %v", err)
	} else {
		magic := make([]byte, 4)
		n, err := file.Read(magic)
		file.Close()
		if err != nil {
			log.Errorf("Error reading magic bytes: %v", err)
		} else {
			log.Infof("Read %d magic bytes: [0x%02x 0x%02x 0x%02x 0x%02x]", n, magic[0], magic[1], magic[2], magic[3])
		}
	}

	if isExcel || ext == ".xlsx" || ext == ".xls" {
		log.Info("Detected Excel file, using Excel parser")
		return parseExcel(csvFilePath)
	}

	log.Info("Using CSV parser")
	return parseCSVFile(csvFilePath)
}

func isExcelFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first 4 bytes to check for ZIP signature (Excel files are ZIP archives)
	magic := make([]byte, 4)
	_, err = file.Read(magic)
	if err != nil {
		return false, err
	}

	// Check for ZIP magic bytes: PK\x03\x04
	return magic[0] == 0x50 && magic[1] == 0x4B && magic[2] == 0x03 && magic[3] == 0x04, nil
}

func parseExcel(filePath string) ([]SharedListSourceEntry, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}

	// Get the first sheet name
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	log.Infof("Reading from sheet: %s", sheetName)

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows from sheet: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows found in Excel file")
	}

	var entries []SharedListSourceEntry

	// Skip header row (first row)
	for i, row := range rows {
		if i == 0 {
			continue // Skip header
		}

		// Excel columns: 0=Developer, 1=(skip), 2=Service Base URL List
		if len(row) < 3 {
			continue
		}

		developerName := strings.TrimSpace(row[0])
		listSource := strings.TrimSpace(row[2])

		// Skip empty entries
		if developerName == "" || listSource == "" {
			continue
		}

		entries = append(entries, SharedListSourceEntry{
			DeveloperName: developerName,
			ListSource:    listSource,
		})
	}

	return entries, nil
}

func parseCSVFile(csvFilePath string) ([]SharedListSourceEntry, error) {
	file, err := os.Open(csvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.Comma = ','
	csvReader.LazyQuotes = true

	// Skip header row
	_, err = csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	var entries []SharedListSourceEntry

	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Warnf("Error reading CSV row: %v", err)
			continue
		}

		// CSV columns: 0=Developer, 1=(skip), 2=Service Base URL List
		if len(rec) < 3 {
			continue
		}

		developerName := strings.TrimSpace(rec[0])
		listSource := strings.TrimSpace(rec[2])

		// Skip empty entries
		if developerName == "" || listSource == "" {
			continue
		}

		entries = append(entries, SharedListSourceEntry{
			DeveloperName: developerName,
			ListSource:    listSource,
		})
	}

	return entries, nil
}

func populateSharedListSources(ctx context.Context, store *postgresql.Store, entries []SharedListSourceEntry) error {
	// Clear existing data
	_, err := store.DB.ExecContext(ctx, "TRUNCATE TABLE shared_list_sources")
	if err != nil {
		return fmt.Errorf("failed to truncate shared_list_sources table: %w", err)
	}

	// Insert new data
	insertStmt := `
		INSERT INTO shared_list_sources (list_source, developer_name, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (list_source, developer_name) DO NOTHING
	`

	successCount := 0
	for _, entry := range entries {
		_, err := store.DB.ExecContext(ctx, insertStmt, entry.ListSource, entry.DeveloperName)
		if err != nil {
			log.Warnf("Failed to insert entry (developer=%s, list_source=%s): %v",
				entry.DeveloperName, entry.ListSource, err)
			continue
		}
		successCount++
	}

	log.Infof("Successfully inserted %d/%d entries", successCount, len(entries))
	return nil
}
