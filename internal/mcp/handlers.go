package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gwork1883/mcp-pprof/internal/pprof"
	"github.com/gwork1883/mcp-pprof/pkg/protocol"
)

// handleParseProfile handles the parse_profile tool
func (s *Server) handleParseProfile(ctx context.Context, args map[string]any) (*protocol.ToolCallResult, error) {
	filePath, ok := args["filePath"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	profileType := pprof.ProfileTypeAuto
	if pt, ok := args["profileType"].(string); ok {
		profileType = pprof.ProfileType(pt)
	}

	output, err := s.pprofWrapper.ParseProfile(filePath, profileType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	jsonOutput, err := s.pprofWrapper.FormatJSON(output)
	if err != nil {
		return nil, fmt.Errorf("failed to format JSON: %w", err)
	}

	return &protocol.ToolCallResult{
		Content: []protocol.ContentBlock{
			{
				Type: "text",
				Text: jsonOutput,
			},
		},
	}, nil
}

// handleTopFunctions handles the top_functions tool
func (s *Server) handleTopFunctions(ctx context.Context, args map[string]any) (*protocol.ToolCallResult, error) {
	filePath, ok := args["filePath"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	topN := 10
	if n, ok := args["topN"].(float64); ok {
		topN = int(n)
	}

	functions, err := s.pprofWrapper.GetTopN(filePath, topN)
	if err != nil {
		return nil, fmt.Errorf("failed to get top functions: %w", err)
	}

	result := map[string]any{
		"filePath": filePath,
		"topN":     topN,
		"results":  functions,
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &protocol.ToolCallResult{
		Content: []protocol.ContentBlock{
			{
				Type: "text",
				Text: string(jsonOutput),
			},
		},
	}, nil
}

// handleGenerateSVG handles the generate_svg tool
func (s *Server) handleGenerateSVG(ctx context.Context, args map[string]any) (*protocol.ToolCallResult, error) {
	filePath, ok := args["filePath"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	focus := ""
	if f, ok := args["focus"].(string); ok {
		focus = f
	}

	ignore := ""
	if i, ok := args["ignore"].(string); ok {
		ignore = i
	}

	svg, err := s.pprofWrapper.GenerateSVG(filePath, focus, ignore)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SVG: %w", err)
	}

	// Truncate SVG if too long for display
	maxLength := 10000
	if len(svg) > maxLength {
		svg = svg[:maxLength] + "\n... (truncated) ..."
	}

	return &protocol.ToolCallResult{
		Content: []protocol.ContentBlock{
			{
				Type: "text",
				Text: svg,
			},
		},
		Metadata: map[string]any{
			"filePath": filePath,
			"format":   "svg",
		},
	}, nil
}

// handleAnalyzePerformance handles the analyze_performance tool
func (s *Server) handleAnalyzePerformance(ctx context.Context, args map[string]any) (*protocol.ToolCallResult, error) {
	filePath, ok := args["filePath"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	focus := "all"
	if f, ok := args["focus"].(string); ok {
		focus = f
	}

	threshold := float64(5)
	if t, ok := args["threshold"].(float64); ok {
		threshold = t
	}

	// Get profile data
	output, err := s.pprofWrapper.ParseProfile(filePath, pprof.ProfileTypeAuto)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	// Analyze and generate suggestions
	result := s.analyzeProfile(output, focus, threshold)

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &protocol.ToolCallResult{
		Content: []protocol.ContentBlock{
			{
				Type: "text",
				Text: string(jsonOutput),
			},
		},
	}, nil
}

// analyzeProfile performs deep analysis of the profile
func (s *Server) analyzeProfile(output *pprof.PprofOutput, focus string, threshold float64) map[string]any {
	result := map[string]any{
		"summary": output.Summary,
	}

	// Identify hotspots
	hotspots := []map[string]any{}
	for _, fn := range output.TopFunctions {
		if fn.Percentage >= threshold {
			hotspot := map[string]any{
				"function":    fn.Name,
				"percentage":  fn.Percentage,
				"cumulative":  fn.Cum,
				"location":    fmt.Sprintf("%s:%d", fn.File, fn.Line),
				"samples":     fn.Samples,
			}
			hotspots = append(hotspots, hotspot)
		}
	}
	result["hotspots"] = hotspots

	// Detect bottlenecks
	bottlenecks := s.detectBottlenecks(output.TopFunctions)
	if len(bottlenecks) > 0 {
		result["bottlenecks"] = bottlenecks
	}

	// Generate suggestions
	suggestions := s.generateSuggestions(output.TopFunctions, output.Summary.ProfileType)
	if len(suggestions) > 0 {
		result["optimizationSuggestions"] = suggestions
	}

	return result
}

// detectBottlenecks detects potential bottlenecks
func (s *Server) detectBottlenecks(functions []pprof.FunctionInfo) []map[string]any {
	bottlenecks := []map[string]any{}

	for _, fn := range functions {
		// Detect I/O related bottlenecks
		if strings.Contains(fn.Name, "syscall") || 
		   strings.Contains(fn.Name, "net.") ||
		   strings.Contains(fn.Name, "File.") ||
		   strings.Contains(fn.Name, "database") {
			bottlenecks = append(bottlenecks, map[string]any{
				"function": fn.Name,
				"type":     "I/O",
				"impact":   s.getImpactLevel(fn.Percentage),
				"percentage": fn.Percentage,
				"suggestion": "Consider using caching or batching to reduce I/O operations",
			})
		}

		// Detect memory allocation bottlenecks
		if strings.Contains(fn.Name, "malloc") ||
		   strings.Contains(fn.Name, "gc") ||
		   strings.Contains(fn.Name, "new") {
			bottlenecks = append(bottlenecks, map[string]any{
				"function": fn.Name,
				"type":     "Memory",
				"impact":   s.getImpactLevel(fn.Percentage),
				"percentage": fn.Percentage,
				"suggestion": "Consider object pooling or reducing allocation frequency",
			})
		}

		// Detect lock related bottlenecks
		if strings.Contains(fn.Name, "Lock") ||
		   strings.Contains(fn.Name, "Mutex") ||
		   strings.Contains(fn.Name, "sync.") {
			bottlenecks = append(bottlenecks, map[string]any{
				"function": fn.Name,
				"type":     "Concurrency",
				"impact":   s.getImpactLevel(fn.Percentage),
				"percentage": fn.Percentage,
				"suggestion": "Consider reducing lock contention or using lock-free data structures",
			})
		}
	}

	return bottlenecks
}

// generateSuggestions generates optimization suggestions
func (s *Server) generateSuggestions(functions []pprof.FunctionInfo, profileType pprof.ProfileType) []map[string]any {
	suggestions := []map[string]any{}

	// Count by function type
	var runtimeCount, gcCount, ioCount int

	for _, fn := range functions {
		if strings.HasPrefix(fn.Name, "runtime.") {
			runtimeCount++
		}
		if strings.Contains(fn.Name, "gc") || strings.Contains(fn.Name, "GC") {
			gcCount++
		}
		if strings.Contains(fn.Name, "net.") || strings.Contains(fn.Name, "File") {
			ioCount++
		}
	}

	// Add suggestions based on analysis
	if gcCount > 3 {
		suggestions = append(suggestions, map[string]any{
			"priority":          "High",
			"area":              "Garbage Collection",
			"suggestion":       "High GC overhead detected. Consider reducing allocations.",
			"estimatedImprovement": "20-40%",
		})
	}

	if ioCount > 2 {
		suggestions = append(suggestions, map[string]any{
			"priority":          "Medium",
			"area":              "I/O Operations",
			"suggestion":       "Multiple I/O operations in hot path. Consider batching or async I/O.",
			"estimatedImprovement": "15-25%",
		})
	}

	if profileType == pprof.ProfileTypeCPU && runtimeCount > 5 {
		suggestions = append(suggestions, map[string]any{
			"priority":          "Low",
			"area":              "Runtime Overhead",
			"suggestion":       "Significant time in runtime. Consider reviewing algorithm complexity.",
			"estimatedImprovement": "10-20%",
		})
	}

	return suggestions
}

// getImpactLevel returns the impact level based on percentage
func (s *Server) getImpactLevel(percentage float64) string {
	if percentage > 20 {
		return "High"
	}
	if percentage > 10 {
		return "Medium"
	}
	return "Low"
}

// handleCompareProfiles handles the compare_profiles tool
func (s *Server) handleCompareProfiles(ctx context.Context, args map[string]any) (*protocol.ToolCallResult, error) {
	baseFile, ok := args["baseFile"].(string)
	if !ok || baseFile == "" {
		return nil, fmt.Errorf("baseFile is required")
	}

	compareFile, ok := args["compareFile"].(string)
	if !ok || compareFile == "" {
		return nil, fmt.Errorf("compareFile is required")
	}

	diff, err := s.pprofWrapper.CompareProfiles(baseFile, compareFile)
	if err != nil {
		return nil, fmt.Errorf("failed to compare profiles: %w", err)
	}

	result := map[string]any{
		"baseFile":    baseFile,
		"compareFile": compareFile,
		"diff":        diff,
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &protocol.ToolCallResult{
		Content: []protocol.ContentBlock{
			{
				Type: "text",
				Text: string(jsonOutput),
			},
		},
	}, nil
}

// handleListCallers handles the list_callers tool
func (s *Server) handleListCallers(ctx context.Context, args map[string]any) (*protocol.ToolCallResult, error) {
	filePath, ok := args["filePath"].(string)
	if !ok || filePath == "" {
		return nil, fmt.Errorf("filePath is required")
	}

	functionName, ok := args["functionName"].(string)
	if !ok || functionName == "" {
		return nil, fmt.Errorf("functionName is required")
	}

	maxDepth := 10
	if md, ok := args["maxDepth"].(float64); ok {
		maxDepth = int(md)
	}

	output, err := s.pprofWrapper.ListCallers(filePath, functionName)
	if err != nil {
		return nil, fmt.Errorf("failed to list callers: %w", err)
	}

	result := map[string]any{
		"function": functionName,
		"filePath": filePath,
		"maxDepth": maxDepth,
		"callers":  output,
	}

	jsonOutput, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &protocol.ToolCallResult{
		Content: []protocol.ContentBlock{
			{
				Type: "text",
				Text: string(jsonOutput),
			},
		},
	}, nil
}
