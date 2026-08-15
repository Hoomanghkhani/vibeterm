package ssh

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"

	"vibeterm/internal/models"
)

// SFTPManager provides remote file management over SSH
type SFTPManager struct{}

// NewSFTPManager creates a new SFTP manager
func NewSFTPManager() *SFTPManager {
	return &SFTPManager{}
}

// OpenSFTP creates an sftp.Client from an active SSH client
func (m *SFTPManager) OpenSFTP(sshClient *ssh.Client) (*sftp.Client, error) {
	return sftp.NewClient(sshClient)
}

// RemoteFileInfo represents a file or directory on the remote server
type RemoteFileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	IsDir   bool   `json:"isDir"`
	ModTime string `json:"modTime"`
}

// ListDirectory lists files in a remote directory
func (m *SFTPManager) ListDirectory(ctx context.Context, host models.Host, remotePath string) ([]RemoteFileInfo, error) {
	cm := GetClientManager()
	client, err := cm.DialHost(ctx, host)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to start sftp subsystem: %w", err)
	}
	defer sftpClient.Close()

	if remotePath == "" {
		remotePath = "."
	}

	files, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}

	var results []RemoteFileInfo
	for _, f := range files {
		results = append(results, RemoteFileInfo{
			Name:    f.Name(),
			Path:    filepath.Join(remotePath, f.Name()),
			Size:    f.Size(),
			Mode:    f.Mode().String(),
			IsDir:   f.IsDir(),
			ModTime: f.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	return results, nil
}

// DownloadFile transfers a remote file to a local destination
func (m *SFTPManager) DownloadFile(ctx context.Context, host models.Host, remotePath, localPath string) error {
	cm := GetClientManager()
	client, err := cm.DialHost(ctx, host)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// UploadFile transfers a local file to a remote destination
func (m *SFTPManager) UploadFile(ctx context.Context, host models.Host, localPath, remotePath string) error {
	cm := GetClientManager()
	client, err := cm.DialHost(ctx, host)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
