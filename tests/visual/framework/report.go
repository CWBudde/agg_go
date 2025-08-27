// Package framework provides HTML report generation for visual test results.
package framework

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// ReportGenerator creates HTML reports for visual test results.
type ReportGenerator struct {
	OutputDir string
}

// NewReportGenerator creates a new report generator.
func NewReportGenerator(outputDir string) *ReportGenerator {
	return &ReportGenerator{
		OutputDir: outputDir,
	}
}

// GenerateReport creates an HTML report for the test suite results.
func (rg *ReportGenerator) GenerateReport(suite *TestSuite) error {
	// Ensure output directory exists
	if err := os.MkdirAll(rg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create reports directory: %v", err)
	}

	// Generate HTML report
	reportPath := filepath.Join(rg.OutputDir, fmt.Sprintf("%s_report.html", suite.Name))
	file, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer file.Close()

	// Execute template
	tmpl := template.Must(template.New("report").Funcs(reportTemplateFuncs()).Parse(reportTemplate))
	if err := tmpl.Execute(file, suite); err != nil {
		return fmt.Errorf("failed to execute report template: %v", err)
	}

	fmt.Printf("HTML report generated: %s\n", reportPath)
	return nil
}

// reportTemplateFuncs provides helper functions for the HTML template.
func reportTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"relpath": func(path string) string {
			// Make paths relative to the report directory for HTML links
			if path == "" {
				return ""
			}
			return filepath.Base(path)
		},
		"formatPercent": func(part, total int) string {
			if total == 0 {
				return "0%"
			}
			percent := float64(part) * 100.0 / float64(total)
			return fmt.Sprintf("%.1f%%", percent)
		},
		"statusClass": func(result TestResult) string {
			if result.Error != nil {
				return "error"
			} else if result.Passed {
				return "pass"
			} else {
				return "fail"
			}
		},
		"statusText": func(result TestResult) string {
			if result.Error != nil {
				return "ERROR"
			} else if result.Passed {
				return "PASS"
			} else {
				return "FAIL"
			}
		},
		"hasImages": func(result TestResult) bool {
			return result.ReferencePath != "" && result.GeneratedPath != ""
		},
		"hasDiff": func(result TestResult) bool {
			return result.DiffPath != ""
		},
		"formatError": func(err error) string {
			if err == nil {
				return ""
			}
			return strings.ReplaceAll(err.Error(), "\n", "<br>")
		},
		"getStats": func(suite *TestSuite) map[string]int {
			stats := map[string]int{"total": 0, "passed": 0, "failed": 0, "errors": 0}
			for _, result := range suite.Results {
				stats["total"]++
				if result.Error != nil {
					stats["errors"]++
				} else if result.Passed {
					stats["passed"]++
				} else {
					stats["failed"]++
				}
			}
			return stats
		},
	}
}

// HTML template for the visual test report.
const reportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Visual Test Report: {{.Name}}</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .header {
            border-bottom: 2px solid #eee;
            padding-bottom: 20px;
            margin-bottom: 30px;
        }
        .title {
            color: #333;
            margin: 0 0 10px 0;
        }
        .meta {
            color: #666;
            font-size: 14px;
        }
        .stats {
            display: flex;
            gap: 20px;
            margin: 20px 0;
        }
        .stat {
            padding: 10px 20px;
            border-radius: 4px;
            text-align: center;
            flex: 1;
        }
        .stat.total { background: #e3f2fd; color: #1976d2; }
        .stat.passed { background: #e8f5e8; color: #2e7d2e; }
        .stat.failed { background: #ffebee; color: #c62828; }
        .stat.errors { background: #fff3e0; color: #f57c00; }
        .stat-number {
            font-size: 24px;
            font-weight: bold;
            display: block;
        }
        .stat-label {
            font-size: 12px;
            text-transform: uppercase;
        }
        .test-result {
            border: 1px solid #ddd;
            margin-bottom: 20px;
            border-radius: 8px;
            overflow: hidden;
        }
        .test-header {
            padding: 15px 20px;
            display: flex;
            align-items: center;
            gap: 15px;
            cursor: pointer;
            user-select: none;
        }
        .test-header:hover {
            background: #fafafa;
        }
        .test-header.pass { border-left: 4px solid #4caf50; }
        .test-header.fail { border-left: 4px solid #f44336; }
        .test-header.error { border-left: 4px solid #ff9800; }
        .status-badge {
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: bold;
            text-transform: uppercase;
        }
        .status-badge.pass { background: #4caf50; color: white; }
        .status-badge.fail { background: #f44336; color: white; }
        .status-badge.error { background: #ff9800; color: white; }
        .test-name {
            flex: 1;
            font-weight: 500;
        }
        .test-details {
            padding: 20px;
            border-top: 1px solid #eee;
            background: #fafafa;
            display: none;
        }
        .test-details.expanded {
            display: block;
        }
        .error-message {
            background: #ffebee;
            border-left: 4px solid #f44336;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 4px;
        }
        .comparison-info {
            background: #fff3e0;
            border-left: 4px solid #ff9800;
            padding: 15px;
            margin-bottom: 20px;
            border-radius: 4px;
        }
        .images-container {
            display: grid;
            grid-template-columns: 1fr 1fr 1fr;
            gap: 20px;
            margin-top: 20px;
        }
        .image-section {
            text-align: center;
        }
        .image-section h4 {
            margin: 0 0 10px 0;
            color: #555;
            font-size: 14px;
            text-transform: uppercase;
        }
        .image-section img {
            max-width: 100%;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .toggle-indicator {
            font-size: 18px;
            transition: transform 0.2s;
        }
        .toggle-indicator.expanded {
            transform: rotate(90deg);
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            text-align: center;
            color: #666;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">Visual Test Report: {{.Name}}</h1>
            <div class="meta">
                Generated: {{.StartTime}} | Duration: {{.Duration}}
            </div>
            
            {{$stats := getStats .}}
            <div class="stats">
                <div class="stat total">
                    <span class="stat-number">{{$stats.total}}</span>
                    <span class="stat-label">Total</span>
                </div>
                <div class="stat passed">
                    <span class="stat-number">{{$stats.passed}}</span>
                    <span class="stat-label">Passed</span>
                </div>
                <div class="stat failed">
                    <span class="stat-number">{{$stats.failed}}</span>
                    <span class="stat-label">Failed</span>
                </div>
                <div class="stat errors">
                    <span class="stat-number">{{$stats.errors}}</span>
                    <span class="stat-label">Errors</span>
                </div>
            </div>
        </div>

        <div class="results">
            {{range .Results}}
            <div class="test-result">
                <div class="test-header {{statusClass .}}" onclick="toggleDetails('{{.Name}}')">
                    <span class="status-badge {{statusClass .}}">{{statusText .}}</span>
                    <span class="test-name">{{.Name}}</span>
                    {{if or .Error (not .Passed) (hasImages .)}}
                    <span class="toggle-indicator" id="toggle-{{.Name}}">▶</span>
                    {{end}}
                </div>
                
                {{if or .Error (not .Passed) (hasImages .)}}
                <div class="test-details" id="details-{{.Name}}">
                    {{if .Error}}
                    <div class="error-message">
                        <strong>Error:</strong><br>
                        {{formatError .Error}}
                    </div>
                    {{else if and .Comparison (not .Passed)}}
                    <div class="comparison-info">
                        <strong>Comparison Results:</strong><br>
                        Different pixels: {{.Comparison.DifferentPixels}} / {{.Comparison.TotalPixels}} ({{formatPercent .Comparison.DifferentPixels .Comparison.TotalPixels}})<br>
                        Maximum difference: {{.Comparison.MaxDifference}} / 255<br>
                        Average difference: {{printf "%.2f" .Comparison.AverageDifference}}
                    </div>
                    {{end}}
                    
                    {{if hasImages .}}
                    <div class="images-container">
                        <div class="image-section">
                            <h4>Reference</h4>
                            <img src="{{relpath .ReferencePath}}" alt="Reference image for {{.Name}}">
                        </div>
                        <div class="image-section">
                            <h4>Generated</h4>
                            <img src="{{relpath .GeneratedPath}}" alt="Generated image for {{.Name}}">
                        </div>
                        {{if hasDiff .}}
                        <div class="image-section">
                            <h4>Difference</h4>
                            <img src="{{relpath .DiffPath}}" alt="Difference image for {{.Name}}">
                        </div>
                        {{end}}
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>
            {{end}}
        </div>

        <div class="footer">
            Generated by AGG Go Visual Testing Framework
        </div>
    </div>

    <script>
        function toggleDetails(testName) {
            const details = document.getElementById('details-' + testName);
            const toggle = document.getElementById('toggle-' + testName);
            
            if (details && toggle) {
                const isExpanded = details.classList.contains('expanded');
                if (isExpanded) {
                    details.classList.remove('expanded');
                    toggle.classList.remove('expanded');
                } else {
                    details.classList.add('expanded');
                    toggle.classList.add('expanded');
                }
            }
        }
    </script>
</body>
</html>`
