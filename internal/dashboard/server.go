package dashboard

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Run represents one date-stamped audit execution folder.
type Run struct {
	FolderName string    `json:"folder_name"`
	ExecutedAt time.Time `json:"executed_at"`
	Files      []string  `json:"files"`
}

// Server is the dashboard HTTP server.
type Server struct {
	OutputDir string
	Addr      string
}

// New creates a Server backed by the given output directory.
func New(outputDir, addr string) *Server {
	return &Server{OutputDir: outputDir, Addr: addr}
}

// Serve starts the HTTP server and opens the default browser.
func (s *Server) Serve() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/runs", s.handleRuns)
	mux.HandleFunc("/api/content/", s.handleContent)
	mux.HandleFunc("/", s.handleUI)

	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("dashboard listen %s: %w", s.Addr, err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	host := l.Addr().(*net.TCPAddr).IP.String()
	localURL := fmt.Sprintf("http://localhost:%d", port)

	fmt.Printf("[oxaudit] Dashboard: %s\n", localURL)

	// When bound to all interfaces, also print LAN addresses for sharing.
	if host == "0.0.0.0" {
		for _, ip := range lanIPs() {
			fmt.Printf("          Network:    http://%s:%d\n", ip, port)
		}
	}
	fmt.Printf("          Output dir: %s\n          Press Ctrl+C to stop.\n", s.OutputDir)

	openBrowser(localURL)
	return http.Serve(l, mux)
}

func (s *Server) listRuns() ([]Run, error) {
	entries, err := os.ReadDir(s.OutputDir)
	if err != nil {
		return nil, err
	}

	var runs []Run
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		t, err := time.Parse("2006-01-02T150405Z", name)
		if err != nil {
			continue // skip non-run folders
		}
		run := Run{
			FolderName: name,
			ExecutedAt: t,
		}

		// Preferred tab order.
		tabOrder := map[string]int{
			"ACTION-PLAN.md":        0,
			"llm-digest.md":         1,
			"findings.md":           2,
			"remediation-backlog.md": 3,
			"findings.csv":          4,
			"findings.jsonl":        5,
		}

		// ACTION-PLAN.md in run root.
		if _, err := os.Stat(filepath.Join(s.OutputDir, name, "ACTION-PLAN.md")); err == nil {
			run.Files = append(run.Files, "ACTION-PLAN.md")
		}

		// Files under exports/.
		exportsDir := filepath.Join(s.OutputDir, name, "exports")
		if es, err := os.ReadDir(exportsDir); err == nil {
			for _, f := range es {
				if !f.IsDir() {
					run.Files = append(run.Files, "exports/"+f.Name())
				}
			}
		}

		sort.SliceStable(run.Files, func(i, j int) bool {
			oi := tabOrder[filepath.Base(run.Files[i])]
			oj := tabOrder[filepath.Base(run.Files[j])]
			if oi != oj {
				return oi < oj
			}
			return run.Files[i] < run.Files[j]
		})

		runs = append(runs, run)
	}

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].ExecutedAt.After(runs[j].ExecutedAt)
	})
	return runs, nil
}

func (s *Server) handleRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.listRuns()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(runs)
}

func (s *Server) handleContent(w http.ResponseWriter, r *http.Request) {
	// Path: /api/content/{folder}/{...file}
	rest := strings.TrimPrefix(r.URL.Path, "/api/content/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	folder, filePath := parts[0], parts[1]
	if strings.Contains(folder, "..") || strings.Contains(filePath, "..") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	http.ServeFile(w, r, filepath.Join(s.OutputDir, folder, filePath))
}

func (s *Server) handleUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML)) //nolint:errcheck
}

// lanIPs returns the machine's non-loopback IPv4 addresses.
func lanIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return ips
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return
	}
	cmd.Start() //nolint:errcheck
}
