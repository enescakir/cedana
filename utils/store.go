package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
)

// Abstraction for storing and retreiving checkpoints
type Store interface {
	GetCheckpoint(string) (*string, error) // returns filepath to downloaded chekcpoint
	PushCheckpoint(filepath string) error
	ListCheckpoints() (*[]CheckpointMeta, error) // fix
}

type CheckpointMeta struct {
	ID       string
	Name     string
	Bucket   string
	ModTime  time.Time
	Size     uint64
	Checksum string
}

type S3Store struct {
	logger *zerolog.Logger
}

func (s *S3Store) GetCheckpoint() (*string, error) {
	return nil, nil
}

func (s *S3Store) PushCheckpoint(filepath string) error {
	return nil
}

type UploadResponse struct {
	UploadID  string `json:"upload_id"`
	PartSize  int64  `json:"part_size"`
	PartCount int64  `json:"part_count"`
}

// For pushing and pulling from a cedana managed endpoint
type CedanaStore struct {
	logger *zerolog.Logger
	cfg    *Config
	url    string
}

func NewCedanaStore(cfg *Config) *CedanaStore {
	logger := GetLogger()
	url := "https://" + cfg.Connection.CedanaUrl
	return &CedanaStore{
		logger: &logger,
		cfg:    cfg,
		url:    url,
	}
}

// TODO NR - unimplemented stubs for now
func (cs *CedanaStore) ListCheckpoints() (*[]CheckpointMeta, error) {
	return nil, nil
}

func (cs *CedanaStore) GetCheckpoint(cid string) (*string, error) {
	url := cs.url + "/checkpoint/" + cid
	downloadPath := "checkpoint.tar"
	file, err := os.Create(downloadPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	httpClient := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cs.cfg.Connection.CedanaUser))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unexpected status code: %v", resp.Status)
	}

	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return nil, err
	}

	return &downloadPath, nil
}

func (cs *CedanaStore) PushCheckpoint(filepath string) error {
	return nil
}

func (cs *CedanaStore) CreateMultiPartUpload(fullSize int64) (UploadResponse, string, error) {
	var uploadResp UploadResponse

	cid := uuid.New().String()

	data := struct {
		Name     string `json:"name"`
		FullSize int64  `json:"full_size"`
		PartSize int    `json:"part_size"`
	}{
		// TODO BS Need to get TaskID properly...
		Name:     "test",
		FullSize: fullSize,
		PartSize: 0,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return uploadResp, "", err
	}

	httpClient := &http.Client{}
	url := cs.url + "/checkpoint/" + cid + "/upload"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return uploadResp, "", err
	}

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cs.cfg.Connection.CedanaUser))

	resp, err := httpClient.Do(req)
	if err != nil {
		return uploadResp, "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return uploadResp, "", fmt.Errorf("unexpected status code: %v", resp.Status)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return uploadResp, "", err
	}

	cs.logger.Info().Msgf("response body: %s", string(respBody))

	// Parse the JSON response into the struct
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		fmt.Println("Error parsing JSON response:", err)
		return uploadResp, "", err
	}

	return uploadResp, cid, nil
}

// expecting part size and part count from server, but not getting it..
func (cs *CedanaStore) StartMultiPartUpload(cid string, uploadResp *UploadResponse, checkpointPath string) error {
	binaryOfFile, err := os.ReadFile(checkpointPath)
	if err != nil {
		fmt.Println("Error reading zip file:", err)
		return err
	}

	chunkSize := uploadResp.PartSize

	numOfParts := uploadResp.PartCount

	for i := 0; i < int(numOfParts); i++ {
		start := i * int(chunkSize)
		end := (i + 1) * int(chunkSize)
		if end > len(binaryOfFile) {
			end = len(binaryOfFile)
		}

		partData := binaryOfFile[start:end]

		buffer := bytes.NewBuffer(partData)

		httpClient := &http.Client{}
		url := cs.url + "/checkpoint/" + cid + "/upload/" + uploadResp.UploadID + "/part/" + fmt.Sprintf("%d", i+1)

		req, err := http.NewRequest("PUT", url, buffer)
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Transfer-Encoding", "chunked")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cs.cfg.Connection.CedanaUser))

		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("unexpected status code: %v", resp.Status)
		}

		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		fmt.Printf("Response: %s\n", respBody)

		cs.logger.Debug().Msgf("Part %d: Size = %d bytes\n", i+1, len(partData))
	}

	return nil
}
func (cs *CedanaStore) CompleteMultiPartUpload(uploadResp UploadResponse, cid string) error {
	httpClient := &http.Client{}
	url := cs.url + "/checkpoint/" + cid + "/upload/" + uploadResp.UploadID + "/complete"

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cs.cfg.Connection.CedanaUser))

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("unexpected status code: %v", resp.Status)
	}

	defer resp.Body.Close()
	return nil
}

type MockStore struct {
	fs     *afero.Afero // we can use an in-memory store for testing
	logger *zerolog.Logger
}

func (ms *MockStore) GetCheckpoint() (*string, error) {
	// gets a mock checkpoint from the local filesystem - useful for testing
	return nil, nil
}

func (ms *MockStore) PushCheckpoint(filepath string) error {
	// pushes a mock checkpoint to the local filesystem
	return nil
}

func (ms *MockStore) ListCheckpoints() (*[]CheckpointMeta, error) {
	return nil, nil
}
