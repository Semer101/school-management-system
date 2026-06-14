package helpers

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dutchcoders/go-clamd"
)

// ClamAVScanner wraps the ClamAV client for virus scanning
type ClamAVScanner struct {
	client *clamd.Clamd
}

// NewClamAVScanner creates a new ClamAV scanner instance
// It connects to ClamAV via TCP (localhost:3310) by default
// Can be configured via CLAMAV_ADDRESS environment variable
func NewClamAVScanner() *ClamAVScanner {
	address := os.Getenv("CLAMAV_ADDRESS")
	if address == "" {
		// Default to TCP connection on localhost
		address = "tcp://127.0.0.1:3310"
	}

	return &ClamAVScanner{
		client: clamd.NewClamd(address),
	}
}

// ScanFile scans a file for viruses using ClamAV
// Returns true if the file is clean, false if infected or error occurred
func (s *ClamAVScanner) ScanFile(filePath string) (bool, string, error) {
	response, err := s.client.ScanFile(filePath)
	if err != nil {
		return false, "", fmt.Errorf("failed to scan file: %w", err)
	}

	// Process the scan results
	for result := range response {
		if result.Status == clamd.RES_FOUND {
			// Virus found
			return false, result.Description, nil
		} else if result.Status == clamd.RES_ERROR {
			// Scan error
			return false, result.Description, fmt.Errorf("scan error: %s", result.Description)
		}
		// RES_OK means the file is clean
	}

	return true, "", nil
}

// ─── Inline Content-Based Malware Scanner (no external dependencies) ──────

// suspiciousPatterns lists regex patterns commonly found in malicious files.
// This provides a "lite" antivirus check that works without ClamAV installed.
var suspiciousPatterns = []struct {
	name    string
	pattern *regexp.Regexp
}{
	// PHP shells / backdoors
	{"PHP Shell", regexp.MustCompile(`(?i)(<?php\s*(system|exec|shell_exec|passthru|popen|eval|assert|base64_decode|preg_replace.*\/e))`)},
	{"PHP Backdoor", regexp.MustCompile(`(?i)(<?php\s*\$\_[GET|POST|REQUEST|COOKIE|SERVER|FILES]\s*\[)`)},
	{"PHP File Write", regexp.MustCompile(`(?i)(fwrite|fputs|file_put_contents)\s*\(`)},
	{"PHP Shell Exec", regexp.MustCompile(`(?i)(?:shell_exec|exec\(|system\(|passthru)`)},
	
	// JavaScript malicious patterns (XSS, crypto miners, etc.)
	{"JavaScript eval", regexp.MustCompile(`(?i)eval\s*\(\s*(document\.write|String\.fromCharCode|unescape|atob)`)},
	{"JavaScript obfuscated", regexp.MustCompile(`(?i)(\\x[0-9a-f]{2}){20,}`)},
	
	// HTML smuggling / phishing
	{"Embedded IFrame", regexp.MustCompile(`(?i)<iframe[^>]*src\s*=\s*["']https?://[^"']*["']`)},
	{"Base64 Data URI", regexp.MustCompile(`(?i)data:\s*(text/html|application/x-javascript|application/javascript);base64,`)},
	
	// Executable indicators in forbidden contexts
	{"PE Executable", regexp.MustCompile(`^MZ\x90`)},
	{"ELF Binary", regexp.MustCompile(`^\x7fELF`)},
	
	// Macro / VBA (in Office docs - may appear as plain text in polyglots)
	{"VBA AutoOpen", regexp.MustCompile(`(?i)(AutoOpen|Auto_Open|Workbook_Open|Document_Open)\s*\(`)},
	{"VBA Shell call", regexp.MustCompile(`(?i)(Shell|CreateObject|WScript\.Shell|ADODB\.Stream)`)},
	
	// Python backdoors (sometimes uploaded as .txt)
	{"Python subprocess", regexp.MustCompile(`(?i)(import\s+(os|subprocess|pty|socket)|from\s+(os|subprocess|pty|socket)\s+import)`)},
	{"Python reverse shell", regexp.MustCompile(`(?i)(socket\.connect\(|socket\.socket\(.*AF_INET|subprocess\.call\(.*shell=True)`)},
	
	// Generic script execution attempts
	{"CMD execution", regexp.MustCompile(`(?i)(cmd\.exe|powershell\.exe|wscript\.exe|cscript\.exe|mshta\.exe)`)},

	// Embedded ZIP bombs / suspicious compression
	{"Nested ZIP bomb", regexp.MustCompile(`(\x50\x4B\x03\x04.*){3,}`)},
}

// maxScanSize limits the bytes read for content scanning to prevent DoS via large files.
const maxScanSize = 1 * 1024 * 1024 // 1 MB

// ScanContentLite performs a signature-based scan of file content for malicious patterns.
// This works entirely offline and does not require any external service.
// Returns (clean, description) where clean=false means a threat was detected.
func ScanContentLite(filePath string) (bool, string) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// For PDFs, also scan for embedded JavaScript which is a known attack vector
	if ext == ".pdf" {
		content, err := readFileHead(filePath)
		if err != nil {
			log.Printf("WARNING: Could not read file %s for content scan: %v", filePath, err)
			return true, ""
		}
		// Check for embedded JS in PDF streams
		if bytes.Contains(content, []byte("/JavaScript")) || bytes.Contains(content, []byte("/JS")) {
			return false, "PDF contains embedded JavaScript (possible malicious PDF)"
		}
	}

	// For all file types, check suspicious patterns
	content, err := readFileHead(filePath)
	if err != nil {
		log.Printf("WARNING: Could not read file %s for content scan: %v", filePath, err)
		return true, ""
	}

	// Lowercase for case-insensitive matching
	lowerContent := bytes.ToLower(content)

	for _, sp := range suspiciousPatterns {
		if sp.pattern.Match(lowerContent) || sp.pattern.Match(content) {
			return false, fmt.Sprintf("Suspicious content detected: %s", sp.name)
		}
	}

	return true, ""
}

// readFileHead reads the first maxScanSize bytes of a file.
func readFileHead(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, maxScanSize)
	n, err := f.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}
	return buf[:n], nil
}

// ScanFileSimple performs multi-layered file scanning:
// 1. Content-based malware pattern detection (offline, no external deps)
// 2. ClamAV scanning if available (online, requires clamd daemon)
//
// Returns true if the file is clean (passes all checks or scanner unavailable).
// Returns false if any scanner detects a threat.
func ScanFileSimple(filePath string) bool {
	// Layer 1: Inline content scanning (always runs, no external deps)
	clean, reason := ScanContentLite(filePath)
	if !clean {
		log.Printf("WARNING: Content scan rejected file %s: %s", filePath, reason)
		return false
	}

	// Layer 2: ClamAV scanning (if available)
	scanner := NewClamAVScanner()
	clean, description, err := scanner.ScanFile(filePath)

	if err != nil {
		// Log but don't block upload if ClamAV is unavailable
		log.Printf("INFO: ClamAV scan not available for %s: %v", filePath, err)
		return true // Allow upload - content scan already passed
	}

	if !clean {
		log.Printf("WARNING: Virus detected in file %s: %s", filePath, description)
		return false
	}

	log.Printf("INFO: File %s passed all security checks", filePath)
	return true
}
