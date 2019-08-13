package sshtrans

import (
	"fmt"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
)

func TransViaPassword(
	hostKey string,
	username string,
	password string,
	localFilepath string,
	remoteFilename string,
	remoteDir string,
) error {
	logrus.WithFields(logrus.Fields{
		"local":     localFilepath,
		"remoteDir": remoteDir,
		"message":   "sending...",
	}).Debug("SND")
	if username == "" {
		return fmt.Errorf("missing username")
	}
	if password == "" {
		return fmt.Errorf("missing password")
	}
	if localFilepath == "" {
		return fmt.Errorf("missing localFilepath")
	}

	_, hosts, pubKey, _, _, err := ssh.ParseKnownHosts([]byte(hostKey))
	if err != nil {
		return fmt.Errorf("invalid host key: %v", err)
	}
	if len(hosts) < 1 {
		return fmt.Errorf("invalid host: %v", hosts)
	}
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.FixedHostKey(pubKey),
	}
	conn, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:22", hosts[0]),
		config,
	)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}

	client, err := sftp.NewClient(conn)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	remoteFilepath := filepath.Join(remoteDir, remoteFilename)
	if remoteDir != "" && remoteDir != "./" && remoteDir != "~/" {
		if client.MkdirAll(remoteDir) != nil {
			return fmt.Errorf(
				"failed to create remote directory '%s': %v",
				remoteDir,
				err,
			)
		}
	}
	dstFile, err := client.Create(remoteFilepath)
	if err != nil {
		return fmt.Errorf("failed to create target file: %v", err)
	}
	if client.Chmod(remoteFilepath, os.FileMode(0755)) != nil {
		logrus.WithFields(logrus.Fields{
			"file":    remoteFilepath,
			"message": "failed to chmod",
		}).Error("SND")
	}
	defer dstFile.Close()

	srcFile, err := os.Open(localFilepath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to send file: %v", err)
	}
	logrus.WithFields(logrus.Fields{
		"bytesSent": bytes,
	}).Debug("SND")
	logrus.WithFields(logrus.Fields{
		"local":  localFilepath,
		"remote": remoteDir,
	}).Info("SND")
	return nil
}
