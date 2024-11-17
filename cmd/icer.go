/*
Copyright (c) 2024 mitsu3s

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
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func main() {
	var typeCode uint8
	var code uint8

	// ICER Command
	var icerCmd = &cobra.Command{
		Use:          "icer",
		Short:        shortMessage,
		Long:         longMessage,
		SilenceUsage: true,
	}

	// Add required flags
	icerCmd.Flags().Uint8VarP(&typeCode, "type", "t", 0, "Type for the ICMP (3=unreachable, 5=redirect, 11=time exceeded)")
	icerCmd.Flags().Uint8VarP(&code, "code", "c", 0, "Code for the ICMP")

	// Version Command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show icer version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("icer version", version)
		},
	}

	// Add version command to ICER command
	icerCmd.AddCommand(versionCmd)

	// Execute ICER command
	icerCmd.Execute()
}
