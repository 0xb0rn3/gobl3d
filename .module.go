package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Target       string   `json:"target"`
	Wordlists    []string `json:"wordlists"`
	Extensions   []string `json:"extensions"`
	Threads      int      `json:"threads"`
	Delay        string   `json:"delay"`
	Timeout      string   `json:"timeout"`
	OutputDir    string   `json:"output_dir"`
	SkipSSL      bool     `json:"skip_ssl"`
	FollowRedir  bool     `json:"follow_redirects"`
	UserAgent    string   `json:"user_agent"`
	Proxy        string   `json:"proxy"`
	Cookies      string   `json:"cookies"`
	Headers      []string `json:"headers"`
	StatusCodes  []string `json:"status_codes"`
	ExcludeCodes []string `json:"exclude_codes"`
}

type ScanResult struct {
	URL        string    `json:"url"`
	StatusCode int       `json:"status_code"`
	Size       int       `json:"size"`
	Timestamp  time.Time `json:"timestamp"`
	ScanType   string    `json:"scan_type"`
}

type UltimateGobuster struct {
	config     *Config
	results    []ScanResult
	resultsMux sync.Mutex
	outputDir  string
}

func NewUltimateGobuster(config *Config) *UltimateGobuster {
	// Create output directory
	timestamp := time.Now().Format("20060102_150405")
	outputDir := filepath.Join(config.OutputDir, fmt.Sprintf("gobuster_scan_%s", timestamp))
	os.MkdirAll(outputDir, 0755)

	return &UltimateGobuster{
		config:    config,
		results:   make([]ScanResult, 0),
		outputDir: outputDir,
	}
}

func (ug *UltimateGobuster) logInfo(message string) {
	fmt.Printf("[INFO] %s\n", message)
}

func (ug *UltimateGobuster) logError(message string) {
	fmt.Printf("[ERROR] %s\n", message)
}

func (ug *UltimateGobuster) logSuccess(message string) {
	fmt.Printf("[SUCCESS] %s\n", message)
}

func (ug *UltimateGobuster) runGobusterCommand(args []string, outputFile string, scanType string) error {
	ug.logInfo(fmt.Sprintf("Running: gobuster %s", strings.Join(args, " ")))
	
	cmd := exec.Command("gobuster", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("gobuster command failed: %v", err)
	}

	// Parse results if output file exists
	if outputFile != "" {
		ug.parseResults(outputFile, scanType)
	}

	return nil
}

func (ug *UltimateGobuster) parseResults(filename, scanType string) {
	file, err := os.Open(filename)
	if err != nil {
		ug.logError(fmt.Sprintf("Failed to open results file %s: %v", filename, err))
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "=") {
			continue
		}

		// Parse gobuster output format
		if strings.Contains(line, "(Status:") {
			result := ug.parseGobusterLine(line, scanType)
			if result != nil {
				ug.resultsMux.Lock()
				ug.results = append(ug.results, *result)
				ug.resultsMux.Unlock()
			}
		}
	}
}

func (ug *UltimateGobuster) parseGobusterLine(line, scanType string) *ScanResult {
	// Example: /admin (Status: 200) [Size: 1234]
	re := regexp.MustCompile(`^(.+?)\s+\(Status:\s+(\d+)\)\s+\[Size:\s+(\d+)\]`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) >= 3 {
		url := strings.TrimSpace(matches[1])
		statusCode := 0
		size := 0
		
		fmt.Sscanf(matches[2], "%d", &statusCode)
		if len(matches) >= 4 {
			fmt.Sscanf(matches[3], "%d", &size)
		}

		return &ScanResult{
			URL:        url,
			StatusCode: statusCode,
			Size:       size,
			Timestamp:  time.Now(),
			ScanType:   scanType,
		}
	}
	return nil
}

func (ug *UltimateGobuster) buildBaseArgs() []string {
	args := []string{}
	
	if ug.config.Threads > 0 {
		args = append(args, "-t", fmt.Sprintf("%d", ug.config.Threads))
	}
	
	if ug.config.Delay != "" {
		args = append(args, "--delay", ug.config.Delay)
	}
	
	if ug.config.Timeout != "" {
		args = append(args, "--timeout", ug.config.Timeout)
	}
	
	if ug.config.SkipSSL {
		args = append(args, "-k")
	}
	
	if ug.config.FollowRedir {
		args = append(args, "-r")
	}
	
	if ug.config.UserAgent != "" {
		args = append(args, "-a", ug.config.UserAgent)
	}
	
	if ug.config.Proxy != "" {
		args = append(args, "--proxy", ug.config.Proxy)
	}
	
	if ug.config.Cookies != "" {
		args = append(args, "-c", ug.config.Cookies)
	}
	
	for _, header := range ug.config.Headers {
		args = append(args, "-H", header)
	}
	
	if len(ug.config.StatusCodes) > 0 {
		args = append(args, "-s", strings.Join(ug.config.StatusCodes, ","))
	}
	
	if len(ug.config.ExcludeCodes) > 0 {
		args = append(args, "-b", strings.Join(ug.config.ExcludeCodes, ","))
	}
	
	return args
}

func (ug *UltimateGobuster) runDirectoryEnum() error {
	ug.logInfo("Starting directory enumeration...")
	
	for i, wordlist := range ug.config.Wordlists {
		outputFile := filepath.Join(ug.outputDir, fmt.Sprintf("directories_%d.txt", i+1))
		
		args := []string{"dir"}
		args = append(args, ug.buildBaseArgs()...)
		args = append(args, "-u", ug.config.Target)
		args = append(args, "-w", wordlist)
		args = append(args, "-o", outputFile)
		
		if len(ug.config.Extensions) > 0 {
			args = append(args, "-x", strings.Join(ug.config.Extensions, ","))
		}
		
		err := ug.runGobusterCommand(args, outputFile, "directory")
		if err != nil {
			ug.logError(fmt.Sprintf("Directory enumeration failed for wordlist %s: %v", wordlist, err))
			continue
		}
		
		ug.logSuccess(fmt.Sprintf("Directory enumeration completed for wordlist %s", wordlist))
	}
	
	return nil
}

func (ug *UltimateGobuster) runSubdomainEnum() error {
	ug.logInfo("Starting subdomain enumeration...")
	
	// Extract domain from target URL
	domain := ug.extractDomain(ug.config.Target)
	if domain == "" {
		return fmt.Errorf("could not extract domain from target: %s", ug.config.Target)
	}
	
	for i, wordlist := range ug.config.Wordlists {
		outputFile := filepath.Join(ug.outputDir, fmt.Sprintf("subdomains_%d.txt", i+1))
		
		args := []string{"dns"}
		args = append(args, ug.buildBaseArgs()...)
		args = append(args, "-d", domain)
		args = append(args, "-w", wordlist)
		args = append(args, "-o", outputFile)
		
		err := ug.runGobusterCommand(args, outputFile, "subdomain")
		if err != nil {
			ug.logError(fmt.Sprintf("Subdomain enumeration failed for wordlist %s: %v", wordlist, err))
			continue
		}
		
		ug.logSuccess(fmt.Sprintf("Subdomain enumeration completed for wordlist %s", wordlist))
	}
	
	return nil
}

func (ug *UltimateGobuster) runVHostEnum() error {
	ug.logInfo("Starting virtual host enumeration...")
	
	for i, wordlist := range ug.config.Wordlists {
		outputFile := filepath.Join(ug.outputDir, fmt.Sprintf("vhosts_%d.txt", i+1))
		
		args := []string{"vhost"}
		args = append(args, ug.buildBaseArgs()...)
		args = append(args, "-u", ug.config.Target)
		args = append(args, "-w", wordlist)
		args = append(args, "-o", outputFile)
		
		err := ug.runGobusterCommand(args, outputFile, "vhost")
		if err != nil {
			ug.logError(fmt.Sprintf("VHost enumeration failed for wordlist %s: %v", wordlist, err))
			continue
		}
		
		ug.logSuccess(fmt.Sprintf("VHost enumeration completed for wordlist %s", wordlist))
	}
	
	return nil
}

func (ug *UltimateGobuster) extractDomain(target string) string {
	// Remove protocol
	target = strings.TrimPrefix(target, "http://")
	target = strings.TrimPrefix(target, "https://")
	
	// Remove port and path
	parts := strings.Split(target, "/")
	domain := parts[0]
	
	if strings.Contains(domain, ":") {
		domain = strings.Split(domain, ":")[0]
	}
	
	return domain
}

func (ug *UltimateGobuster) runRecursiveEnum() error {
	ug.logInfo("Starting recursive enumeration on interesting directories...")
	
	// Get interesting directories from previous scans
	interestingDirs := ug.getInterestingDirectories()
	
	for _, dir := range interestingDirs {
		targetURL := strings.TrimSuffix(ug.config.Target, "/") + dir
		outputFile := filepath.Join(ug.outputDir, fmt.Sprintf("recursive_%s.txt", strings.ReplaceAll(dir, "/", "_")))
		
		// Use first wordlist for recursive scan
		if len(ug.config.Wordlists) > 0 {
			args := []string{"dir"}
			args = append(args, ug.buildBaseArgs()...)
			args = append(args, "-u", targetURL)
			args = append(args, "-w", ug.config.Wordlists[0])
			args = append(args, "-o", outputFile)
			
			if len(ug.config.Extensions) > 0 {
				args = append(args, "-x", strings.Join(ug.config.Extensions, ","))
			}
			
			err := ug.runGobusterCommand(args, outputFile, "recursive")
			if err != nil {
				ug.logError(fmt.Sprintf("Recursive enumeration failed for %s: %v", dir, err))
				continue
			}
			
			ug.logSuccess(fmt.Sprintf("Recursive enumeration completed for %s", dir))
		}
	}
	
	return nil
}

func (ug *UltimateGobuster) getInterestingDirectories() []string {
	interesting := []string{}
	interestingKeywords := []string{"admin", "api", "backup", "config", "login", "panel", "upload", "files", "docs"}
	
	ug.resultsMux.Lock()
	defer ug.resultsMux.Unlock()
	
	for _, result := range ug.results {
		if result.ScanType == "directory" && result.StatusCode == 200 {
			for _, keyword := range interestingKeywords {
				if strings.Contains(strings.ToLower(result.URL), keyword) {
					interesting = append(interesting, result.URL)
					break
				}
			}
		}
	}
	
	return interesting
}

func (ug *UltimateGobuster) generateReport() error {
	ug.logInfo("Generating comprehensive report...")
	
	// Generate JSON report
	jsonFile := filepath.Join(ug.outputDir, "report.json")
	ug.resultsMux.Lock()
	jsonData, err := json.MarshalIndent(ug.results, "", "  ")
	ug.resultsMux.Unlock()
	
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	
	err = ioutil.WriteFile(jsonFile, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON report: %v", err)
	}
	
	// Generate summary report
	summaryFile := filepath.Join(ug.outputDir, "summary.txt")
	summary := ug.generateSummary()
	
	err = ioutil.WriteFile(summaryFile, []byte(summary), 0644)
	if err != nil {
		return fmt.Errorf("failed to write summary report: %v", err)
	}
	
	ug.logSuccess(fmt.Sprintf("Reports generated in: %s", ug.outputDir))
	return nil
}

func (ug *UltimateGobuster) generateSummary() string {
	ug.resultsMux.Lock()
	defer ug.resultsMux.Unlock()
	
	var summary strings.Builder
	summary.WriteString("=== ULTIMATE GOBUSTER SCAN SUMMARY ===\n\n")
	summary.WriteString(fmt.Sprintf("Target: %s\n", ug.config.Target))
	summary.WriteString(fmt.Sprintf("Scan Date: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	summary.WriteString(fmt.Sprintf("Total Results: %d\n\n", len(ug.results)))
	
	// Group by scan type
	scanTypes := make(map[string][]ScanResult)
	for _, result := range ug.results {
		scanTypes[result.ScanType] = append(scanTypes[result.ScanType], result)
	}
	
	for scanType, results := range scanTypes {
		summary.WriteString(fmt.Sprintf("=== %s RESULTS (%d) ===\n", strings.ToUpper(scanType), len(results)))
		for _, result := range results {
			summary.WriteString(fmt.Sprintf("  %s (Status: %d, Size: %d)\n", result.URL, result.StatusCode, result.Size))
		}
		summary.WriteString("\n")
	}
	
	return summary.String()
}

func (ug *UltimateGobuster) Run() error {
	ug.logInfo("Starting Ultimate Gobuster Bug Bounty Tool...")
	ug.logInfo(fmt.Sprintf("Target: %s", ug.config.Target))
	ug.logInfo(fmt.Sprintf("Output Directory: %s", ug.outputDir))
	
	// Run different enumeration types
	var wg sync.WaitGroup
	
	// Directory enumeration
	wg.Add(1)
	go func() {
		defer wg.Done()
		ug.runDirectoryEnum()
	}()
	
	// Subdomain enumeration
	wg.Add(1)
	go func() {
		defer wg.Done()
		ug.runSubdomainEnum()
	}()
	
	// VHost enumeration
	wg.Add(1)
	go func() {
		defer wg.Done()
		ug.runVHostEnum()
	}()
	
	// Wait for initial scans to complete
	wg.Wait()
	
	// Run recursive enumeration on interesting directories
	ug.runRecursiveEnum()
	
	// Generate reports
	ug.generateReport()
	
	ug.logSuccess("Ultimate Gobuster scan completed!")
	return nil
}

func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	
	return &config, nil
}

func createDefaultConfig() *Config {
	return &Config{
		Target: "http://example.com",
		Wordlists: []string{
			"/usr/share/wordlists/dirb/common.txt",
			"/usr/share/wordlists/SecLists/Discovery/Web-Content/directory-list-2.3-medium.txt",
		},
		Extensions:   []string{"php", "html", "txt", "js", "asp", "aspx"},
		Threads:      30,
		Delay:        "100ms",
		Timeout:      "10s",
		OutputDir:    "./gobuster_results",
		SkipSSL:      true,
		FollowRedir:  true,
		UserAgent:    "Mozilla/5.0 (compatible; UltimateGobuster/1.0)",
		StatusCodes:  []string{"200", "301", "302", "401", "403"},
		ExcludeCodes: []string{"404", "500"},
	}
}

func main() {
	var (
		configFile = flag.String("config", "", "Path to configuration file")
		target     = flag.String("target", "", "Target URL")
		outputDir  = flag.String("output", "./gobuster_results", "Output directory")
		threads    = flag.Int("threads", 30, "Number of threads")
		genConfig  = flag.Bool("gen-config", false, "Generate default configuration file")
	)
	flag.Parse()
	
	if *genConfig {
		config := createDefaultConfig()
		data, _ := json.MarshalIndent(config, "", "  ")
		err := ioutil.WriteFile("config.json", data, 0644)
		if err != nil {
			log.Fatal("Failed to create config file:", err)
		}
		fmt.Println("Default configuration created: config.json")
		return
	}
	
	var config *Config
	var err error
	
	if *configFile != "" {
		config, err = loadConfig(*configFile)
		if err != nil {
			log.Fatal("Failed to load config:", err)
		}
	} else {
		config = createDefaultConfig()
	}
	
	// Override config with command line flags
	if *target != "" {
		config.Target = *target
	}
	if *outputDir != "" {
		config.OutputDir = *outputDir
	}
	if *threads > 0 {
		config.Threads = *threads
	}
	
	if config.Target == "" || config.Target == "http://example.com" {
		log.Fatal("Please provide a target URL using -target flag or config file")
	}
	
	// Create and run the ultimate gobuster
	ug := NewUltimateGobuster(config)
	err = ug.Run()
	if err != nil {
		log.Fatal("Scan failed:", err)
	}
}
