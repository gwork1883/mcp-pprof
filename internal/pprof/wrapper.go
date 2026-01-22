package pprof

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ProfileType represents the type of profile
type ProfileType string

const (
	ProfileTypeAuto      ProfileType = "auto"
	ProfileTypeCPU       ProfileType = "cpu"
	ProfileTypeHeap      ProfileType = "heap"
	ProfileTypeBlock     ProfileType = "block"
	ProfileTypeMutex     ProfileType = "mutex"
	ProfileTypeGoroutine ProfileType = "goroutine"
)

// ProfileSummary represents a profile summary
type ProfileSummary struct {
	ProfileType ProfileType `json:"profileType"`
	TotalSamples int64       `json:"totalSamples"`
	TimeRange    string      `json:"timeRange,omitempty"`
	SampleRate   int         `json:"sampleRate,omitempty"`
}

// FunctionInfo represents function information
type FunctionInfo struct {
	Name        string  `json:"name"`
	Samples     int64   `json:"samples"`
	Percentage  float64 `json:"percentage"`
	Flat        float64 `json:"flat"`
	Cum         float64 `json:"cum"`
	File        string  `json:"file,omitempty"`
	Line        int     `json:"line,omitempty"`
}

// PprofOutput represents parsed pprof output
type PprofOutput struct {
	Summary     ProfileSummary   `json:"summary"`
	TopFunctions []FunctionInfo  `json:"topFunctions,omitempty"`
	RawText     string          `json:"rawText,omitempty"`
}

// Wrapper wraps go tool pprof functionality
type Wrapper struct {
	toolPath string
}

// NewWrapper creates a new pprof wrapper
func NewWrapper() *Wrapper {
	toolPath, _ := exec.LookPath("go")
	return &Wrapper{
		toolPath: toolPath,
	}
}

// ParseProfile parses a pprof file and returns structured data
func (w *Wrapper) ParseProfile(filePath string, profileType ProfileType) (*PprofOutput, error) {
	// First, get text output
	output, err := w.runPprof("-text", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	result := &PprofOutput{
		RawText: output,
	}

	// Parse the output
	result.Summary = w.parseSummary(filePath, profileType)
	result.TopFunctions = w.parseTextOutput(output)

	return result, nil
}

// GetTopN returns top N functions
func (w *Wrapper) GetTopN(filePath string, n int) ([]FunctionInfo, error) {
	output, err := w.runPprof("-top", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get top functions: %w", err)
	}

	functions := w.parseTextOutput(output)
	if len(functions) > n {
		functions = functions[:n]
	}

	// Assign ranks
	for i := range functions {
		functions[i].Samples = int64(functions[i].Flat * 10000) // Approximate
	}

	return functions, nil
}

// GenerateSVG generates SVG output
func (w *Wrapper) GenerateSVG(filePath string, focus, ignore string) (string, error) {
	args := []string{"-svg"}
	
	if focus != "" {
		args = append(args, "-focus", focus)
	}
	if ignore != "" {
		args = append(args, "-ignore", ignore)
	}
	
	args = append(args, filePath)
	
	return w.runPprof(args...)
}

// ListCallers lists the callers of a function
func (w *Wrapper) ListCallers(filePath, functionName string) (string, error) {
	return w.runPprof("-list", functionName, filePath)
}

// runPprof executes go tool pprof with given arguments
func (w *Wrapper) runPprof(args ...string) (string, error) {
	if w.toolPath == "" {
		return "", fmt.Errorf("go tool not found")
	}
	
	fullArgs := append([]string{"tool", "pprof"}, args...)
	cmd := exec.Command(w.toolPath, fullArgs...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("pprof command failed: %w, stderr: %s", err, stderr.String())
	}
	
	return stdout.String(), nil
}

// parseSummary parses summary information
func (w *Wrapper) parseSummary(filePath string, profileType ProfileType) ProfileSummary {
	// Try to get file info
	if _, err := os.Stat(filePath); err == nil {
		sampleRate := 100
		if profileType == ProfileTypeCPU {
			sampleRate = 100
		}
		
		return ProfileSummary{
			ProfileType: profileType,
			SampleRate:  sampleRate,
		}
	}
	
	return ProfileSummary{
		ProfileType: profileType,
	}
}

// parseTextOutput parses text format pprof output
func (w *Wrapper) parseTextOutput(output string) []FunctionInfo {
	functions := []FunctionInfo{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	// Skip header lines (usually start with empty line, then header line)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		// Check if this looks like header
		if strings.Contains(line, "flat") && strings.Contains(line, "flat%") {
			continue
		}
		
		// Parse the function line
		// Format: flat%   flat%   sum%      cum%    cum%   name
		funcInfo := w.parseFunctionLine(line)
		if funcInfo != nil && funcInfo.Name != "" {
			functions = append(functions, *funcInfo)
		}
	}
	
	return functions
}

// parseFunctionLine parses a single function line
func (w *Wrapper) parseFunctionLine(line string) *FunctionInfo {
	// Remove extra spaces
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil
	}
	
	info := &FunctionInfo{}
	
	// Parse flat percentage
	if flatPct, err := strconv.ParseFloat(strings.TrimSuffix(fields[0], "%"), 64); err == nil {
		info.Flat = flatPct
		info.Percentage = flatPct
	}
	
	// Parse cumulative percentage
	if cnt := len(fields); cnt >= 6 {
		if cumPct, err := strconv.ParseFloat(strings.TrimSuffix(fields[5], "%"), 64); err == nil {
			info.Cum = cumPct
		}
	}
	
	// Function name is typically the last field(s)
	// Find the function name (usually after the last percentage)
	funcNameStart := 6
	if funcNameStart >= len(fields) {
		return nil
	}
	
	name := strings.Join(fields[funcNameStart:], " ")
	info.Name = w.cleanFunctionName(name)
	
	// Try to extract file and line if available in format "file:line"
	if idx := strings.LastIndex(name, ":"); idx > 0 {
		if lineNum, err := strconv.Atoi(name[idx+1:]); err == nil {
			potentialFile := name[:idx]
			if strings.Contains(potentialFile, ".go") {
				info.File = potentialFile
				info.Line = lineNum
				// Remove file:line from function name
				info.Name = w.cleanFunctionName(name[:idx])
			}
		}
	}
	
	return info
}

// cleanFunctionName removes common prefixes/suffixes from function names
func (w *Wrapper) cleanFunctionName(name string) string {
	// Remove common patterns
	name = strings.TrimSpace(name)
	
	// Remove any file path prefix
	if idx := strings.LastIndex(name, "/"); idx > 0 {
		name = name[idx+1:]
	}
	
	return name
}

// ParseTopOutput parses go tool pprof -top output
func (w *Wrapper) ParseTopOutput(output string) (*PprofOutput, error) {
	result := &PprofOutput{
		RawText:     output,
		TopFunctions: w.parseTextOutput(output),
	}
	
	// Calculate total samples from first line if available
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Total:") {
			// Extract total if available
			re := regexp.MustCompile(`Total:\s*([\d.]+)`)
			if matches := re.FindStringSubmatch(line); len(matches) > 1 {
				if total, err := strconv.ParseFloat(matches[1], 64); err == nil {
					result.Summary.TotalSamples = int64(total * 100)
				}
			}
		}
	}
	_ = lines // Use lines to avoid unused variable warning
	
	return result, nil
}

// GetRawText returns raw text output from pprof
func (w *Wrapper) GetRawText(filePath string) (string, error) {
	return w.runPprof("-text", filePath)
}

// CompareProfiles compares two profiles
func (w *Wrapper) CompareProfiles(baseFile, compareFile string) (string, error) {
	return w.runPprof("-base", baseFile, compareFile)
}

// FormatJSON formats output as JSON
func (w *Wrapper) FormatJSON(output *PprofOutput) (string, error) {
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}
