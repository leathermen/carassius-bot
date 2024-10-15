package helpers

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/thanhpk/randstr"
)

func DownloadFile(url string, ext ...string) (*os.File, error) {
	resp, err := http.Get(url) //nolint:noctx

	if err != nil {
		return nil, fmt.Errorf("failed to get url <%s>: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status at file downloading: %d", resp.StatusCode)
	}

	filename := os.TempDir() + randstr.String(24)

	if len(ext) > 0 {
		filename += "."
		filename += ext[0]
	}

	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create tmp file: %w", err)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to copy remote file content to tmp file: %w", err)
	}

	_, _ = file.Seek(0, 0)

	return file, nil
}
