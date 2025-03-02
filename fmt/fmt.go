package fmt

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cli/safeexec"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

const (
	envFileKey         = "FILE"
	defaultShell       = "sh"
	defaultPlaceholder = "?"
)

type Hcl struct {
	shell       string
	placeholder string
	fmtcmds     map[string]string
	bin         string
}

func New(fmtcmds map[string]string) *Hcl {
	return &Hcl{
		shell:       defaultShell,
		placeholder: defaultPlaceholder,
		fmtcmds:     fmtcmds,
	}
}

func (h *Hcl) Format(in []byte) ([]byte, error) {
	bin, err := safeexec.LookPath(h.shell)
	if err != nil {
		return nil, fmt.Errorf("failed to find %s: %w", h.shell, err)
	}
	h.bin = bin
	f, diags := hclwrite.ParseConfig(in, "tmp.hcl", hcl.Pos{})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL: %w", diags)
	}

	if err := h.formatStringLiterals(f.Body()); err != nil {
		return nil, fmt.Errorf("failed to format HCL: %w", err)
	}

	formatted := hclwrite.Format(f.Bytes())

	return formatted, nil
}

func (h *Hcl) formatStringLiterals(body *hclwrite.Body) error {
	for _, attr := range body.Attributes() {
		var (
			fmtcmd     string
			v          string
			next       bool
			fieldToken *hclwrite.Token
			valueToken *hclwrite.Token
			trimTokens []*hclwrite.Token
			interp     string
			names      []string
		)
		for _, t := range attr.BuildTokens(nil) {
			if t.Type == hclsyntax.TokenIdent {
				cmd, ok := h.fmtcmds[string(t.Bytes)]
				if ok {
					fieldToken = t
					fmtcmd = cmd
					continue
				}
			}
			if fmtcmd == "" {
				continue
			}

			switch t.Type {
			case hclsyntax.TokenEqual:
				continue
			case hclsyntax.TokenOHeredoc:
				next = true
				continue
			case hclsyntax.TokenCHeredoc:
				var (
					repPairs    []string
					revertPairs []string
				)
				for i, n := range names {
					repPairs = append(revertPairs, n, fmt.Sprintf("hclstr%dRep", i))
					revertPairs = append(repPairs, fmt.Sprintf("hclstr%dRep", i), n)
				}
				rep := strings.NewReplacer(repPairs...)
				revert := strings.NewReplacer(revertPairs...)

				v = rep.Replace(v)
				f, err := os.CreateTemp("", "hclstr")
				if err != nil {
					return fmt.Errorf("failed to create temp file: %w", err)
				}
				if _, err := f.WriteString(v); err != nil {
					return fmt.Errorf("failed to write to temp file: %w", err)
				}
				fmtcmd = strings.ReplaceAll(fmtcmd, h.placeholder, f.Name())
				cmd := exec.Command(h.bin, "-c", fmtcmd)
				cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", envFileKey, f.Name()))
				if out, err := cmd.CombinedOutput(); err != nil {
					return fmt.Errorf("failed to run %q: %s", fmtcmd, string(out))
				}
				b, err := os.ReadFile(f.Name())
				if err != nil {
					return fmt.Errorf("failed to read temp file: %w", err)
				}
				formatted := revert.Replace(indent(string(b), fieldToken.SpacesBefore))

				valueToken.Type = hclsyntax.TokenStringLit
				valueToken.Bytes = []byte(formatted)
				valueToken.SpacesBefore = 0

				for _, t := range trimTokens {
					t.Type = hclsyntax.TokenStringLit
					t.Bytes = nil
				}

				fmtcmd = ""
				v = ""
				names = nil
				continue
			}

			switch t.Type {
			case hclsyntax.TokenTemplateInterp:
				interp = string(t.Bytes)
			case hclsyntax.TokenTemplateSeqEnd:
				interp += string(t.Bytes)
				names = append(names, interp)
				interp = ""
			case hclsyntax.TokenStringLit:
			case hclsyntax.TokenTemplateControl:
				// directive does not support
				fmtcmd = ""
				v = ""
				interp = ""
				names = nil
				trimTokens = nil
				continue
			default: // case hclsyntax.TokenIdent, hclsyntax.TokenDot:
				interp += string(t.Bytes)
			}
			v += string(t.Bytes)

			if next {
				valueToken = t
				next = false
			} else {
				trimTokens = append(trimTokens, t)
			}
		}
	}

	for _, block := range body.Blocks() {
		if err := h.formatStringLiterals(block.Body()); err != nil {
			return err
		}
	}
	return nil
}

func indent(s string, n int) string {
	lines := strings.Split(s, "\n")
	indentation := strings.Repeat(" ", n)
	for i, line := range lines {
		lines[i] = indentation + line
	}
	return strings.Join(lines, "\n")
}
