package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Valid services in the Smart Food Court Management System
var validServices = []string{
	"ai-analytics-service",
	"api-gateway",
	"dining-iot-service",
	"manager-dashboard",
	"order-kitchen-service",
	"staff-login",
	"user-dashboard",
	"user-service",
	"wallet-service",
}

func main() {
	// Load root .env file if it exists
	_ = godotenv.Overload(".env")
	_ = godotenv.Overload("../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Fallback check: try reading from user-service/.env which has the auth DB
		if err := godotenv.Overload("user-service/.env"); err == nil {
			dbURL = os.Getenv("DB_DSN")
		}
	}

	if dbURL == "" {
		log.Println("Warning: DATABASE_URL or DB_DSN environment variable not found.")
		log.Println("Commands database operations (migrate, flushall, createadmin, etc.) may fail.")
		log.Println("Ensure DATABASE_URL is set in a root .env file, or configure your database connections.")
	}

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := strings.ToLower(os.Args[1])
	switch command {
	case "migrate":
		runMigrate(dbURL)
	case "makemigrations":
		if len(os.Args) < 4 {
			fmt.Println("Error: makemigrations requires service name and description.")
			fmt.Println("Usage: go run manage.go makemigrations <service_name> <description>")
			fmt.Println("Example: go run manage.go makemigrations user-service \"create_users_table\"")
			return
		}
		runMakeMigrations(os.Args[2], os.Args[3])
	case "flushall":
		runFlushAll(dbURL)
	case "createadmin":
		runCreateAdmin(dbURL)
	case "createstaff":
		staffDbURL := ""
		if err := godotenv.Overload("staff-login/.env"); err == nil {
			staffDbURL = os.Getenv("DB_DSN")
		}
		if staffDbURL == "" {
			log.Fatal("Error: Could not read DB_DSN from staff-login/.env. Please ensure the file exists and is configured.")
		}
		runCreateStaff(staffDbURL)
	case "updatepass":
		runUpdatePass(dbURL)
	case "updateemail":
		runUpdateEmail(dbURL)
	case "stop":
		runStopServices()
	case "start", "run":
		runStartServices()
	case "help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printHelp()
	}
}

func printHelp() {
	fmt.Println("\n==================================================")
	fmt.Println("Smart Food Court Database Management Tool")
	fmt.Println("==================================================")
	fmt.Println("Usage:")
	fmt.Println("  go run manage.go <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  migrate                              Apply all pending SQL migrations across all services")
	fmt.Println("  makemigrations <service> <desc>      Generate a new SQL migration template for a specific service")
	fmt.Println("  flushall                             Truncate all data tables (clears DB but keeps schemas)")
	fmt.Println("  createadmin                          Create a new admin user in the users table")
	fmt.Println("  createstaff                          Create a new staff member (MANAGER/CHEF/ADMIN) in staff table")
	fmt.Println("  updatepass                           Update an admin password in the users table")
	fmt.Println("  updateemail                          Update an admin email address in the users table")
	fmt.Println("  start                                Start all microservices in separate console windows (aliased: run)")
	fmt.Println("  stop                                 Stop all microservices by terminating port bindings")
	fmt.Println("  help                                 Show this help screen")
	fmt.Println("\nValid Services:")
	for _, s := range validServices {
		fmt.Printf("  - %s\n", s)
	}
	fmt.Println("==================================================")
}

func connectDB(dbURL string) *gorm.DB {
	if dbURL == "" {
		log.Fatal("Error: Database connection URL is empty. Please set DATABASE_URL.")
	}
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	return db
}

// ----------------------------------------------------
// MIGRATE COMMAND
// ----------------------------------------------------
func runMigrate(dbURL string) {
	// GORM AutoMigrate is skipped for now until models are implemented in the Go microservices.
	runGormAutoMigrate(dbURL)

	fmt.Println("\n[Migrate] Connecting to database to apply SQL migrations...")
	db := connectDB(dbURL)

	// Create global schema_migrations table if not exists
	err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			service VARCHAR(100),
			version VARCHAR(255),
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (service, version)
		)
	`).Error
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// Query already applied migrations
	type MigrationKey struct {
		Service string
		Version string
	}
	var appliedList []MigrationKey
	err = db.Raw("SELECT service, version FROM schema_migrations").Scan(&appliedList).Error
	if err != nil {
		log.Fatalf("Failed to query applied migrations: %v", err)
	}

	appliedMap := make(map[string]bool)
	for _, m := range appliedList {
		key := fmt.Sprintf("%s:%s", m.Service, m.Version)
		appliedMap[key] = true
	}

	// Scan through all microservices for migrations
	rootPath := findWorkspaceRoot()
	migrationsApplied := 0

	for _, service := range validServices {
		serviceMigDir := filepath.Join(rootPath, service, "migrations")
		if _, err := os.Stat(serviceMigDir); os.IsNotExist(err) {
			continue // No migrations directory for this service
		}

		files, err := os.ReadDir(serviceMigDir)
		if err != nil {
			log.Printf("Warning: Failed to read migrations for service %s: %v", service, err)
			continue
		}

		var sqlFiles []string
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
				sqlFiles = append(sqlFiles, file.Name())
			}
		}

		// Sort migrations alphabetically so they execute in order (e.g. 0001, 0002)
		sort.Strings(sqlFiles)

		for _, sqlFile := range sqlFiles {
			key := fmt.Sprintf("%s:%s", service, sqlFile)
			if appliedMap[key] {
				continue // Already applied
			}

			filePath := filepath.Join(serviceMigDir, sqlFile)
			queryBytes, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatalf("Failed to read migration file %s: %v", filePath, err)
			}

			query := string(queryBytes)
			if strings.TrimSpace(query) == "" {
				// Empty file, mark as applied directly
				fmt.Printf("[%s] Marking empty migration as applied: %s\n", service, sqlFile)
			} else {
				fmt.Printf("[%s] Applying migration: %s...\n", service, sqlFile)
				err = db.Transaction(func(tx *gorm.DB) error {
					if err := tx.Exec(query).Error; err != nil {
						return err
					}
					return tx.Exec("INSERT INTO schema_migrations (service, version) VALUES (?, ?)", service, sqlFile).Error
				})
				if err != nil {
					log.Fatalf("Fatal: Migration failed in file %s: %v", sqlFile, err)
				}
			}

			migrationsApplied++
		}
	}

	if migrationsApplied == 0 {
		fmt.Println("Database is already up to date. No pending migrations.")
	} else {
		fmt.Printf("Success: Applied %d migration(s) successfully.\n", migrationsApplied)
	}
}

// ----------------------------------------------------
// MAKEMIGRATIONS COMMAND
// ----------------------------------------------------
func runMakeMigrations(service, description string) {
	service = strings.ToLower(service)

	// Validate service name
	isValid := false
	for _, s := range validServices {
		if s == service {
			isValid = true
			break
		}
	}

	if !isValid {
		fmt.Printf("Error: '%s' is not a valid service name.\n", service)
		fmt.Println("Valid services are:", strings.Join(validServices, ", "))
		return
	}

	// Clean description to be safe for filenames
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	cleanDesc := reg.ReplaceAllString(description, "_")
	cleanDesc = strings.ToLower(cleanDesc)

	rootPath := findWorkspaceRoot()
	serviceMigDir := filepath.Join(rootPath, service, "migrations")

	// Ensure service migrations folder exists
	if err := os.MkdirAll(serviceMigDir, 0755); err != nil {
		log.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Scan existing files to find the next sequence index
	files, err := os.ReadDir(serviceMigDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	maxSeq := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			parts := strings.SplitN(file.Name(), "_", 2)
			if len(parts) > 0 {
				seq, err := strconv.Atoi(parts[0])
				if err == nil && seq > maxSeq {
					maxSeq = seq
				}
			}
		}
	}

	nextSeq := maxSeq + 1
	filename := fmt.Sprintf("%04d_%s.sql", nextSeq, cleanDesc)
	filePath := filepath.Join(serviceMigDir, filename)

	// Create migration template contents
	template := fmt.Sprintf(`-- Migration: %s
-- Created At: %s
-- Description: %s

-- Write your UP SQL migration queries here:
-- E.g. ALTER TABLE table_name ADD COLUMN column_name data_type;

`, filename, time.Now().Format(time.RFC3339), description)

	err = os.WriteFile(filePath, []byte(template), 0644)
	if err != nil {
		log.Fatalf("Failed to write migration template file: %v", err)
	}

	relPath := filepath.Join(service, "migrations", filename)
	fmt.Printf("\nSuccess: Created migration file template at:\n%s\n", relPath)
}

// ----------------------------------------------------
// FLUSHALL COMMAND
// ----------------------------------------------------
func runFlushAll(dbURL string) {
	fmt.Println("\nWARNING: This will truncate all data tables in the database!")
	fmt.Println("This clears all records but keeps database tables and schemas intact.")
	fmt.Print("Are you sure you want to proceed? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}

	confirm := strings.TrimSpace(strings.ToLower(input))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("Operation cancelled.")
		return
	}

	fmt.Println("\n[FlushAll] Connecting to database...")
	db := connectDB(dbURL)

	fmt.Println("[FlushAll] Truncating all tables in public schema...")
	flushSQL := `
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				-- Skip schema_migrations table so we don't lose migration history
				IF r.tablename != 'schema_migrations' THEN
					EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' RESTART IDENTITY CASCADE;';
				END IF;
			END LOOP;
		END $$;
	`

	err = db.Exec(flushSQL).Error
	if err != nil {
		log.Fatalf("Fatal: Database flush failed: %v", err)
	}

	fmt.Println("Success: Database tables truncated and primary key sequences reset successfully.")
}

func findWorkspaceRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}

	dir := cwd
	for {
		workFile := filepath.Join(dir, "go.work")
		if _, err := os.Stat(workFile); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "."
}

// ----------------------------------------------------
// CREATEADMIN COMMAND (Using Bcrypt)
// ----------------------------------------------------
func runCreateAdmin(dbURL string) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Smart Food Court Admin SQL Query Generator (Go - Bcrypt)")
	fmt.Println(strings.Repeat("=", 60))

	reader := bufio.NewReader(os.Stdin)

	// Prompt for email
	fmt.Print("Enter Admin Email: ")
	emailInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading email: %v\n", err)
		os.Exit(1)
	}
	email := strings.ToLower(strings.TrimSpace(emailInput))
	if email == "" {
		fmt.Println("Error: Email cannot be empty.")
		os.Exit(1)
	}

	// Prompt for password
	password, err := readPassword(reader, "Enter Admin Password: ")
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}
	if password == "" {
		fmt.Println("Error: Password cannot be empty.")
		os.Exit(1)
	}

	// Confirm password
	confirmPassword, err := readPassword(reader, "Confirm Admin Password: ")
	if err != nil {
		fmt.Printf("Error reading confirmation: %v\n", err)
		os.Exit(1)
	}
	if password != confirmPassword {
		fmt.Println("Error: Passwords do not match.")
		os.Exit(1)
	}

	// Generate Bcrypt hash
	fmt.Println("\nHashing password using Bcrypt...")
	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Generate SQL Statement for users table
	sqlQuery := fmt.Sprintf(`INSERT INTO users (email, password, role, created_at)
VALUES (
    '%s', 
    '%s', 
    'ADMIN', 
    NOW()
);`, email, hashedPassword)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  GENERATED SQL QUERY (Copy and run this in your database client)")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(sqlQuery)
	fmt.Println(strings.Repeat("=", 60))

	filename := "create_admin.sql"
	err = os.WriteFile(filename, []byte(sqlQuery), 0644)
	if err != nil {
		fmt.Printf("\nWarning: Could not save query to file: %v\n", err)
	} else {
		fmt.Printf("\nSaved query to file: %s\n", filename)
	}

	// Execute immediately if requested
	fmt.Print("\nDo you want to execute this query on the database immediately? (y/N): ")
	confirmInput, err := reader.ReadString('\n')
	if err == nil {
		confirm := strings.TrimSpace(strings.ToLower(confirmInput))
		if confirm == "y" || confirm == "yes" {
			fmt.Println("Connecting to database...")
			db := connectDB(dbURL)
			if err := db.Exec(sqlQuery).Error; err != nil {
				fmt.Printf("Error executing query: %v\n", err)
			} else {
				fmt.Println("Success: Query executed and admin user created successfully!")
			}
		}
	}
}

// ----------------------------------------------------
// UPDATEPASS COMMAND (Using Bcrypt)
// ----------------------------------------------------
func runUpdatePass(dbURL string) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Smart Food Court Admin Password Update SQL Generator (Go - Bcrypt)")
	fmt.Println(strings.Repeat("=", 60))

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Admin Email to Update: ")
	emailInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading email: %v\n", err)
		os.Exit(1)
	}
	email := strings.ToLower(strings.TrimSpace(emailInput))
	if email == "" {
		fmt.Println("Error: Email cannot be empty.")
		os.Exit(1)
	}

	password, err := readPassword(reader, "Enter New Admin Password: ")
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}
	if password == "" {
		fmt.Println("Error: Password cannot be empty.")
		os.Exit(1)
	}

	confirmPassword, err := readPassword(reader, "Confirm New Admin Password: ")
	if err != nil {
		fmt.Printf("Error reading confirmation: %v\n", err)
		os.Exit(1)
	}
	if password != confirmPassword {
		fmt.Println("Error: Passwords do not match.")
		os.Exit(1)
	}

	fmt.Println("\nHashing password using Bcrypt...")
	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	sqlQuery := fmt.Sprintf(`UPDATE users 
SET password = '%s' 
WHERE email = '%s';`, hashedPassword, email)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  GENERATED SQL QUERY (Copy and run this in your database client)")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(sqlQuery)
	fmt.Println(strings.Repeat("=", 60))

	filename := "update_admin.sql"
	err = os.WriteFile(filename, []byte(sqlQuery), 0644)
	if err != nil {
		fmt.Printf("\nWarning: Could not save query to file: %v\n", err)
	} else {
		fmt.Printf("\nSaved query to file: %s\n", filename)
	}

	fmt.Print("\nDo you want to execute this query on the database immediately? (y/N): ")
	confirmInput, err := reader.ReadString('\n')
	if err == nil {
		confirm := strings.TrimSpace(strings.ToLower(confirmInput))
		if confirm == "y" || confirm == "yes" {
			fmt.Println("Connecting to database...")
			db := connectDB(dbURL)
			if err := db.Exec(sqlQuery).Error; err != nil {
				fmt.Printf("Error executing query: %v\n", err)
			} else {
				fmt.Println("Success: Query executed and password updated successfully!")
			}
		}
	}
}

// ----------------------------------------------------
// UPDATEEMAIL COMMAND
// ----------------------------------------------------
func runUpdateEmail(dbURL string) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Smart Food Court Admin Email Update SQL Generator (Go)")
	fmt.Println(strings.Repeat("=", 60))

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Current Admin Email: ")
	oldEmailInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading email: %v\n", err)
		os.Exit(1)
	}
	oldEmail := strings.ToLower(strings.TrimSpace(oldEmailInput))
	if oldEmail == "" {
		fmt.Println("Error: Current Email cannot be empty.")
		os.Exit(1)
	}

	fmt.Print("Enter New Admin Email: ")
	newEmailInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading email: %v\n", err)
		os.Exit(1)
	}
	newEmail := strings.ToLower(strings.TrimSpace(newEmailInput))
	if newEmail == "" {
		fmt.Println("Error: New Email cannot be empty.")
		os.Exit(1)
	}

	sqlQuery := fmt.Sprintf(`UPDATE users 
SET email = '%s' 
WHERE email = '%s';`, newEmail, oldEmail)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  GENERATED SQL QUERY (Copy and run this in your database client)")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(sqlQuery)
	fmt.Println(strings.Repeat("=", 60))

	filename := "update_admin_email.sql"
	err = os.WriteFile(filename, []byte(sqlQuery), 0644)
	if err != nil {
		fmt.Printf("\nWarning: Could not save query to file: %v\n", err)
	} else {
		fmt.Printf("\nSaved query to file: %s\n", filename)
	}

	fmt.Print("\nDo you want to execute this query on the database immediately? (y/N): ")
	confirmInput, err := reader.ReadString('\n')
	if err == nil {
		confirm := strings.TrimSpace(strings.ToLower(confirmInput))
		if confirm == "y" || confirm == "yes" {
			fmt.Println("Connecting to database...")
			db := connectDB(dbURL)
			if err := db.Exec(sqlQuery).Error; err != nil {
				fmt.Printf("Error executing query: %v\n", err)
			} else {
				fmt.Println("Success: Query executed and email updated successfully!")
			}
		}
	}
}

// Helper functions
func readPassword(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		bytePassword, err := term.ReadPassword(fd)
		fmt.Println()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(bytePassword)), nil
	}
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func runGormAutoMigrate(dbURL string) {
	// Stub: Dynamic GORM AutoMigrate is bypassed.
	// You can define standard GORM model schemas here once your entities are built.
	fmt.Println("[Migrate] GORM AutoMigrate is currently disabled (models pending design). Using SQL-based migrations.")
}

// ----------------------------------------------------
// CREATESTAFF COMMAND (Using Bcrypt)
// ----------------------------------------------------
func runCreateStaff(dbURL string) {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Smart Food Court Staff SQL Query Generator (Go - Bcrypt)")
	fmt.Println(strings.Repeat("=", 60))

	reader := bufio.NewReader(os.Stdin)

	// Prompt for Name
	fmt.Print("Enter Staff Name (e.g. Rahul Manager): ")
	nameInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading name: %v\n", err)
		os.Exit(1)
	}
	name := strings.TrimSpace(nameInput)
	if name == "" {
		fmt.Println("Error: Name cannot be empty.")
		os.Exit(1)
	}

	// Prompt for Email
	fmt.Print("Enter Staff Email: ")
	emailInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading email: %v\n", err)
		os.Exit(1)
	}
	email := strings.ToLower(strings.TrimSpace(emailInput))
	if email == "" {
		fmt.Println("Error: Email cannot be empty.")
		os.Exit(1)
	}

	// Prompt for Phone
	fmt.Print("Enter Staff Phone: ")
	phoneInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading phone: %v\n", err)
		os.Exit(1)
	}
	phone := strings.TrimSpace(phoneInput)
	if phone == "" {
		fmt.Println("Error: Phone cannot be empty.")
		os.Exit(1)
	}

	// Prompt for Role
	fmt.Print("Enter Staff Role (MANAGER, CHEF, ADMIN): ")
	roleInput, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading role: %v\n", err)
		os.Exit(1)
	}
	role := strings.ToUpper(strings.TrimSpace(roleInput))
	if role != "MANAGER" && role != "CHEF" && role != "ADMIN" {
		fmt.Println("Error: Invalid role. Must be MANAGER, CHEF, or ADMIN.")
		os.Exit(1)
	}

	// Prompt for password
	password, err := readPassword(reader, "Enter Staff Password: ")
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}
	if password == "" {
		fmt.Println("Error: Password cannot be empty.")
		os.Exit(1)
	}

	// Confirm password
	confirmPassword, err := readPassword(reader, "Confirm Staff Password: ")
	if err != nil {
		fmt.Printf("Error reading confirmation: %v\n", err)
		os.Exit(1)
	}
	if password != confirmPassword {
		fmt.Println("Error: Passwords do not match.")
		os.Exit(1)
	}

	// Generate Bcrypt hash
	fmt.Println("\nHashing password using Bcrypt...")
	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Generate SQL Statement for staff_members table
	sqlQuery := fmt.Sprintf(`INSERT INTO staff_members (id, name, email, phone, password, role, is_active, created_at)
VALUES (
    gen_random_uuid(),
    '%s', 
    '%s', 
    '%s',
    '%s', 
    '%s', 
    true,
    NOW()
);`, name, email, phone, hashedPassword, role)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("  GENERATED SQL QUERY (Copy and run this in your database client)")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(sqlQuery)
	fmt.Println(strings.Repeat("=", 60))

	filename := "create_staff.sql"
	err = os.WriteFile(filename, []byte(sqlQuery), 0644)
	if err != nil {
		fmt.Printf("\nWarning: Could not save query to file: %v\n", err)
	} else {
		fmt.Printf("\nSaved query to file: %s\n", filename)
	}

	// Execute immediately if requested
	fmt.Print("\nDo you want to execute this query on the database immediately? (y/N): ")
	confirmInput, err := reader.ReadString('\n')
	if err == nil {
		confirm := strings.TrimSpace(strings.ToLower(confirmInput))
		if confirm == "y" || confirm == "yes" {
			fmt.Println("Connecting to database...")
			db := connectDB(dbURL)
			if err := db.Exec(sqlQuery).Error; err != nil {
				fmt.Printf("Error executing query: %v\n", err)
			} else {
				fmt.Println("Success: Query executed and staff user created successfully!")
			}
		}
	}
}

// ----------------------------------------------------
// STOP COMMAND
// ----------------------------------------------------
func runStopServices() {
	ports := []int{8080, 8081, 8082, 8084, 8086, 8087, 8088}
	fmt.Println("\nStopping all Smart Food Court microservices...")

	pidRegexp := regexp.MustCompile(`\s+(\d+)\s*$`)
	terminatedCount := 0
	killedPIDs := make(map[string]bool)

	for _, port := range ports {
		if runtime.GOOS == "windows" {
			// Windows: Find PID on local port listening
			cmd := exec.Command("cmd", "/c", fmt.Sprintf("netstat -ano | findstr LISTENING | findstr :%d", port))
			output, err := cmd.Output()
			if err != nil {
				// No process running on this port, skip
				continue
			}

			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				matches := pidRegexp.FindStringSubmatch(line)
				if len(matches) > 1 {
					pid := matches[1]
					if killedPIDs[pid] {
						continue // Already killed in this loop
					}

					fmt.Printf("Port %d is occupied by PID %s. Terminating...\n", port, pid)
					
					// Force terminate process
					killCmd := exec.Command("taskkill", "/F", "/PID", pid)
					if err := killCmd.Run(); err != nil {
						fmt.Printf("Failed to terminate process %s: %v\n", pid, err)
					} else {
						fmt.Printf("Successfully terminated process %s on port %d.\n", pid, port)
						terminatedCount++
						killedPIDs[pid] = true
					}
				}
			}
		} else {
			// Unix-based systems (Linux/macOS)
			cmd := exec.Command("lsof", "-t", fmt.Sprintf("-i:%d", port))
			output, err := cmd.Output()
			if err != nil {
				continue
			}
			
			pid := strings.TrimSpace(string(output))
			if pid != "" {
				if killedPIDs[pid] {
					continue
				}

				fmt.Printf("Port %d is occupied by PID %s. Terminating...\n", port, pid)
				killCmd := exec.Command("kill", "-9", pid)
				if err := killCmd.Run(); err != nil {
					fmt.Printf("Failed to terminate process %s: %v\n", pid, err)
				} else {
					fmt.Printf("Successfully terminated process %s on port %d.\n", pid, port)
					terminatedCount++
					killedPIDs[pid] = true
				}
			}
		}
	}

	if terminatedCount == 0 {
		fmt.Println("No active services detected on port ranges.")
	} else {
		fmt.Printf("Success: Terminated %d microservice process(es).\n", terminatedCount)
	}
}

// ----------------------------------------------------
// START/RUN COMMAND
// ----------------------------------------------------
func runStartServices() {
	services := []struct {
		Name string
		Port int
		Path string
	}{
		{Name: "api-gateway", Port: 8080, Path: "api-gateway"},
		{Name: "user-service", Port: 8081, Path: "user-service"},
		{Name: "wallet-service", Port: 8082, Path: "wallet-service"},
		{Name: "order-kitchen-service", Port: 8084, Path: "order-kitchen-service"},
		{Name: "staff-login", Port: 8086, Path: "staff-login"},
		{Name: "manager-dashboard", Port: 8087, Path: "manager-dashboard"},
		{Name: "user-dashboard", Port: 8088, Path: "user-dashboard"},
	}

	fmt.Println("\nChecking ports and launching microservices...")
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	for _, svc := range services {
		// Check if port is already active
		if isPortInUse(svc.Port) {
			fmt.Printf("[-] Service '%s' is already running on port %d.\n", svc.Name, svc.Port)
			continue
		}

		fmt.Printf("[+] Launching '%s' on port %d in a new terminal window...\n", svc.Name, svc.Port)
		absPath := filepath.Join(cwd, svc.Path)

		if runtime.GOOS == "windows" {
			// Windows: Spawn in a new PowerShell console window, clearing inherited PORT variable
			psCmd := fmt.Sprintf("Start-Process powershell -ArgumentList '-NoExit', '-Command', '$Host.UI.RawUI.WindowTitle = ''%s''; $env:PORT=$null; go run ./cmd/server/main.go' -WorkingDirectory '%s'", svc.Name, absPath)
			execCmd := exec.Command("powershell", "-Command", psCmd)
			if err := execCmd.Run(); err != nil {
				fmt.Printf("Failed to launch service '%s': %v\n", svc.Name, err)
			}
		} else {
			// Unix-based systems: Run in background and redirect logs to log file, clearing inherited PORT variable
			logFile := filepath.Join(absPath, svc.Name+".log")
			shellCmd := fmt.Sprintf("cd '%s' && unset PORT && go run ./cmd/server/main.go > '%s' 2>&1 &", absPath, logFile)
			execCmd := exec.Command("sh", "-c", shellCmd)
			if err := execCmd.Run(); err != nil {
				fmt.Printf("Failed to launch service '%s': %v\n", svc.Name, err)
			} else {
				fmt.Printf("Launched '%s' in background. Logs redirected to: %s\n", svc.Name, logFile)
			}
		}
	}
	fmt.Println("All launches triggered successfully!")
}

// Helper: Checks if a local port is in use using standard TCP listeners
func isPortInUse(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return true
	}
	ln.Close()
	return false
}
