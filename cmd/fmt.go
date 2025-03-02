/*
Copyright © 2025 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	hclfmt "github.com/k1LoW/hclstr/fmt"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	cmds  []string
	check bool
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [FILE ...]",
	Short: "Format HCL files and string literals in HCL files",
	Long: `Format HCL files and string literals in HCL files.
For each string literal field, a different formatter can be specified.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmdcmds := map[string]string{}
		for _, c := range cmds {
			splitted := strings.Split(c, ":")
			if len(splitted) != 2 {
				return fmt.Errorf("invalid format command: %s", c)
			}
			fmdcmds[splitted[0]] = splitted[1]
		}
		cf := hclfmt.New(fmdcmds)
		eg := new(errgroup.Group)
		diffExists := false
		for _, fp := range args {
			if _, err := os.Stat(fp); err != nil {
				return err
			}
			eg.Go(func() error {
				b, err := os.ReadFile(fp)
				if err != nil {
					return err
				}
				formatted, err := cf.Format(b)
				if err != nil {
					return fmt.Errorf("%s: %w", fp, err)
				}
				if check && !bytesEqual(b, formatted) {
					diffExists = true
					fmt.Println(fp)
					return nil
				}
				if err := os.WriteFile(fp, formatted, 0600); err != nil {
					return err
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return err
		}
		if check && diffExists {
			os.Exit(1)
		}
		return nil
	},
}

func bytesEqual(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i, b := range b1 {
		if b != b2[i] {
			return false
		}
	}
	return true
}

func init() {
	rootCmd.AddCommand(fmtCmd)
	fmtCmd.Flags().StringSliceVarP(&cmds, "field", "f", []string{}, "format command for string literal field in HCL files. e.g. 'Expr:deno fmt ${FILE} --ext js'")
	fmtCmd.Flags().BoolVarP(&check, "check", "", false, "exit with status non-zero if not formatted")
}
