package mcom

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
)

// Print prints pdf contents on the specified printer.
// For a cups printers, printerName means "queue name".
// Run this func with verbose flag to see details.
func Print(ctx context.Context, printerName string, r io.Reader) error {
	path := randomPath("pdf")

	if err := writeFile(path, r); err != nil {
		return err
	}
	defer func() {
		if err := deleteFile(path); err != nil {
			commonsCtx.Logger(ctx).Warn("failed to delete temp file", zap.String("file", path))
		}
	}()

	return printFile(path, printerName)
}

func randomPath(extensionName string) string {
	return fmt.Sprintf("./%d.%s", time.Now().UnixNano(), extensionName)
}

func writeFile(path string, r io.Reader) error {
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, r); err != nil {
		return err
	}

	if err = outFile.Sync(); err != nil {
		return err
	}
	return nil
}

// printFile prints specified file on specified printer.
//
// If you use cups as printing system, printerName is "queue name"
// of a cups printer.
func printFile(path string, printerName string) error {
	cmd := exec.Command("lp", "-d", printerName, path)

	_, err := cmd.Output()
	if err == nil {
		return nil
	}

	ok, err2 := hasPrinter(printerName)
	if err2 != nil {
		return fmt.Errorf("%s. %s", err, err2)
	}

	if ok {
		return err
	}

	return fmt.Errorf("printer not found")
}

func hasPrinter(name string) (bool, error) {
	cmd := exec.Command("lpstat", "-v")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check if printer exists: %s", err)
	}

	outContent := string(out)
	lines := strings.Split(outContent, "\n")
	for _, line := range lines {
		if name == parsePrinterName(line) {
			return true, nil
		}
	}

	return false, nil
}

func parsePrinterName(str string) string {
	str = strings.TrimPrefix(str, "device for ")
	parts := strings.Split(str, ":")
	return parts[0]
}

func deleteFile(path string) error {
	return os.Remove(path)
}
