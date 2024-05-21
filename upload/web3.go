package upload

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/starlinglab/integrity-v2/config"
	"github.com/starlinglab/integrity-v2/util"
)

func uploadWeb3(space string, cidPaths []string) error {
	conf := config.GetConfig()

	if _, err := os.Stat(conf.Bins.W3); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("w3 (w3cli) not found at configured path, may not be installed: %s", conf.Bins.W3)
	}

	// Set space
	cmd := exec.Command(conf.Bins.W3, "space", "use", space)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n%s\n", output)
		return fmt.Errorf("w3 (w3cli) failed to use space, see output above if any. Error was: %w", err)
	}

	// Warn for what util.GetCAR might do
	fmt.Fprintln(os.Stderr,
		"warning: the whole file may be loaded into memory to create a CAR file for upload")

	for i, cidPath := range cidPaths {
		// Use anon func to allow for safe idiomatic usage of `defer`
		err := func() error {
			fmt.Printf("Uploading %d of %d...\n", i+1, len(cidPaths))

			// First create a temporary CAR file. Using a CAR file forces web3.storage to
			// use the same CIDs as us instead of generating them in their own different
			// way (--cid-version=1 --chunker=size-1048576).

			tmpF, err := os.CreateTemp("", "upload_")
			if err != nil {
				return fmt.Errorf("error creating temp CAR file: %w", err)
			}
			defer tmpF.Close()
			defer os.Remove(tmpF.Name())

			cidF, err := os.Open(cidPath)
			if err != nil {
				return fmt.Errorf("error opening CID file: %w", err)
			}
			defer cidF.Close()

			car, err := util.GetCAR(cidF)
			if err != nil {
				return fmt.Errorf("error calculating CAR data: %w", err)
			}

			// Make sure CID hasn't changed
			if car.Root().String() != filepath.Base(cidPath) {
				return fmt.Errorf(
					"CAR CID doesn't match file CID: %s != %s",
					car.Root().String(), filepath.Base(cidPath),
				)
			}

			if err := car.Write(tmpF); err != nil {
				return fmt.Errorf("error writing temp CAR file: %w", err)
			}
			tmpF.Close() // Flush for w3

			// Now upload that CAR file
			cmd = exec.Command(conf.Bins.W3, "up", "--car", tmpF.Name())
			output, err = cmd.CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\n%s\n", output)
				return fmt.Errorf("w3 (w3cli) failed to upload, see output above if any. Error was: %w", err)
			}

			err = logUploadWithAA(filepath.Base(cidPath), "web3", "web3.storage", space)
			if err != nil {
				return fmt.Errorf("error logging upload to AuthAttr: %w", err)
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	fmt.Println("Done.")
	return nil
}