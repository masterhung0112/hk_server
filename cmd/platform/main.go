// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/masterhung0112/hk_server/v5/utils/fileutils"
)

func main() {
	// Print angry message to use mattermost command directly
	fmt.Println(`
------------------------------------ ERROR ------------------------------------------------
The platform binary has been deprecated, please switch to using the new mattermost binary.
The platform binary will be removed in a future version.
-------------------------------------------------------------------------------------------
	`)

	// Execve the real MM binary
	args := os.Args
	args[0] = "hkserver"
	args = append(args, "--platform")

	realMattermost := fileutils.FindFile("hkserver")
	if realMattermost == "" {
		realMattermost = fileutils.FindFile("bin/hkserver")
	}

	if realMattermost == "" {
		fmt.Println("Could not start HungKnow, use the hkserver command directly: failed to find mattermost")
	} else if err := syscall.Exec(realMattermost, args, nil); err != nil {
		fmt.Printf("Could not start HungKnow, use the hkserver command directly: %s\n", err.Error())
	}
}
