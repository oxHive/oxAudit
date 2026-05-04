package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Progress tracks and prints pipeline execution progress.
type Progress struct {
	w     io.Writer
	total int
}

// New creates a Progress writer that outputs to stdout.
func New(totalSteps int) *Progress {
	return &Progress{w: os.Stdout, total: totalSteps}
}

func (p *Progress) printf(format string, args ...any) {
	fmt.Fprintf(p.w, format, args...)
}

// PipelineStart prints the opening line.
func (p *Progress) PipelineStart() {
	p.printf("[oxaudit] Starting full audit pipeline\n")
}

// StepStart prints that a step is beginning.
func (p *Progress) StepStart(n int, name string) {
	p.printf("[%d/%d] %-10s running...\n", n, p.total, name)
}

// SubItem prints a sub-item line indented under the current step.
func (p *Progress) SubItem(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	p.printf("  → %s\n", msg)
}

// StepDone prints that a step completed successfully.
func (p *Progress) StepDone(n int, name string, elapsed time.Duration, detail string) {
	if detail != "" {
		p.printf("[%d/%d] %-10s done  (%s, %s)\n", n, p.total, name, fmtDuration(elapsed), detail)
	} else {
		p.printf("[%d/%d] %-10s done  (%s)\n", n, p.total, name, fmtDuration(elapsed))
	}
}

// StepFailed prints that a step failed.
func (p *Progress) StepFailed(n int, name string, err error) {
	p.printf("[%d/%d] %-10s FAILED: %v\n", n, p.total, name, err)
}

// Summary prints the final bordered summary block.
func (p *Progress) Summary(fields [][2]string) {
	border := strings.Repeat("━", 42)
	p.printf("\n%s\n", border)
	for _, f := range fields {
		p.printf("  %-18s %s\n", f[0]+":", f[1])
	}
	p.printf("%s\n", border)
}

func fmtDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
