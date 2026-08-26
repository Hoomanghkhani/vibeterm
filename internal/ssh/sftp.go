package ssh

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

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

	// Resolve absolute path or pwd
	if remotePath == "." {
		if pwd, err := sftpClient.Getwd(); err == nil && pwd != "" {
			remotePath = pwd
		}
	}

	files, err := sftpClient.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}

	var results []RemoteFileInfo
	for _, f := range files {
		results = append(results, RemoteFileInfo{
			Name:    f.Name(),
			Path:    path.Join(remotePath, f.Name()),
			Size:    f.Size(),
			Mode:    f.Mode().String(),
			IsDir:   f.IsDir(),
			ModTime: f.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	return results, nil
}

// UploadData writes byte content directly to a remote destination (for web drag & drop)
func (m *SFTPManager) UploadData(ctx context.Context, host models.Host, remotePath string, data []byte) error {
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

	dstFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.Write(data)
	return err
}

// ReadFileContent reads remote text file content
func (m *SFTPManager) ReadFileContent(ctx context.Context, host models.Host, remotePath string, maxBytes int64) (string, error) {
	cm := GetClientManager()
	client, err := cm.DialHost(ctx, host)
	if err != nil {
		return "", err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return "", err
	}
	defer sftpClient.Close()

	srcFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	if maxBytes <= 0 {
		maxBytes = 512 * 1024 // 512KB limit for preview
	}

	buf := make([]byte, maxBytes)
	n, err := srcFile.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	return string(buf[:n]), nil
}

// DeleteRemoteFile deletes a remote file or folder
func (m *SFTPManager) DeleteRemoteFile(ctx context.Context, host models.Host, remotePath string) error {
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

	stat, err := sftpClient.Stat(remotePath)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return sftpClient.RemoveDirectory(remotePath)
	}
	return sftpClient.Remove(remotePath)
}

// CreateRemoteDir creates a new directory
func (m *SFTPManager) CreateRemoteDir(ctx context.Context, host models.Host, remotePath string) error {
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

	return sftpClient.MkdirAll(remotePath)
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
