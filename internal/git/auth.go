package git

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	ssh2 "golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var (
	useAgent          = true
	passphraseCache   = make(map[string]string)
	passphraseCacheMu sync.RWMutex
	skipPassphrase    = false
)

func SetUseAgent(use bool) {
	useAgent = use
}

func GetUseAgent() bool {
	return useAgent
}

func SetSkipPassphrase(skip bool) {
	skipPassphrase = skip
}

func GetSkipPassphrase() bool {
	return skipPassphrase
}

type ErrPassphraseRequired struct {
	KeyPath string
}

func (e *ErrPassphraseRequired) Error() string {
	return fmt.Sprintf("passphrase required for SSH key: %s", e.KeyPath)
}

func IsPassphraseRequiredError(err error) bool {
	_, ok := err.(*ErrPassphraseRequired)
	return ok
}

func getAuthMethod(url string) (transport.AuthMethod, error) {
	ep, err := transport.NewEndpoint(url)
	if err != nil {
		return nil, err
	}

	if ep.Protocol == "https" || ep.Protocol == "http" {
		return getHTTPAuth()
	}

	if ep.Protocol == "ssh" {
		return getSSHAuth()
	}

	return nil, nil
}

func getHTTPAuth() (transport.AuthMethod, error) {
	username := os.Getenv("GIT_USERNAME")
	password := os.Getenv("GIT_PASSWORD")
	token := os.Getenv("GIT_TOKEN")

	if token != "" {
		return &http.BasicAuth{
			Username: "x-access-token",
			Password: token,
		}, nil
	}

	if username != "" && password != "" {
		return &http.BasicAuth{
			Username: username,
			Password: password,
		}, nil
	}

	return nil, nil
}

func getSSHAuth() (transport.AuthMethod, error) {
	if useAgent {
		auth, err := ssh.NewSSHAgentAuth("git")
		if err == nil {
			auth.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
			return auth, nil
		}
	}

	sshKeyPath := findSSHKey()
	if sshKeyPath != "" {
		auth, err := tryLoadSSHKeyWithPassphrase(sshKeyPath)
		if err != nil {
			return nil, err
		}
		return auth, nil
	}

	return nil, fmt.Errorf("SSH URL detected but no SSH agent or key found. Please either:\n  1. Start your SSH agent\n  2. Set SSH_KEY_PATH environment variable\n  3. Use HTTPS URL instead: https://github.com/user/repo.git\n  4. Set GIT_TOKEN for HTTPS authentication")
}

func findSSHKey() string {
	sshKeyPath := os.Getenv("SSH_KEY_PATH")
	if sshKeyPath != "" {
		if _, err := os.Stat(sshKeyPath); err == nil {
			return sshKeyPath
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	defaultKeys := []string{
		filepath.Join(homeDir, ".ssh", "id_ed25519"),
		filepath.Join(homeDir, ".ssh", "id_rsa"),
		filepath.Join(homeDir, ".ssh", "id_ecdsa"),
	}

	for _, keyPath := range defaultKeys {
		if _, err := os.Stat(keyPath); err == nil {
			return keyPath
		}
	}

	return ""
}

func tryLoadSSHKeyWithPassphrase(keyPath string) (*ssh.PublicKeys, error) {
	if envPass := os.Getenv("SSH_PASSPHRASE"); envPass != "" {
		auth, err := ssh.NewPublicKeysFromFile("git", keyPath, envPass)
		if err == nil {
			auth.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
			return auth, nil
		}
	}

	auth, err := ssh.NewPublicKeysFromFile("git", keyPath, "")
	if err == nil {
		auth.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
		return auth, nil
	}

	if !isPassphraseError(err) {
		return nil, err
	}

	if skipPassphrase {
		return nil, &ErrPassphraseRequired{KeyPath: keyPath}
	}

	if cached, ok := getCachedPassphrase(keyPath); ok {
		auth, err := ssh.NewPublicKeysFromFile("git", keyPath, cached)
		if err == nil {
			auth.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
			return auth, nil
		}
	}

	passphrase, err := promptPassphrase(keyPath)
	if err != nil {
		return nil, err
	}

	auth, err = ssh.NewPublicKeysFromFile("git", keyPath, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key (wrong passphrase?): %w", err)
	}
	auth.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
	return auth, nil
}

func isPassphraseError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "cannot decode encrypted private key") ||
		strings.Contains(errStr, "this private key is passphrase protected") ||
		strings.Contains(errStr, "decryption password incorrect") ||
		strings.Contains(errStr, "x509: decryption password incorrect")
}

func getCachedPassphrase(keyPath string) (string, bool) {
	passphraseCacheMu.RLock()
	defer passphraseCacheMu.RUnlock()
	pass, ok := passphraseCache[keyPath]
	return pass, ok
}

func setCachedPassphrase(keyPath, passphrase string) {
	passphraseCacheMu.Lock()
	defer passphraseCacheMu.Unlock()
	passphraseCache[keyPath] = passphrase
}

func promptPassphrase(keyPath string) (string, error) {
	if cached, ok := getCachedPassphrase(keyPath); ok {
		return cached, nil
	}

	fmt.Printf("Enter passphrase for SSH key %s: ", keyPath)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		passBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", fmt.Errorf("failed to read passphrase: %w", err)
		}
		passphrase := string(passBytes)
		setCachedPassphrase(keyPath, passphrase)
		return passphrase, nil
	}

	reader := bufio.NewReader(os.Stdin)
	passphrase, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read passphrase: %w", err)
	}
	passphrase = strings.TrimSpace(passphrase)
	setCachedPassphrase(keyPath, passphrase)
	return passphrase, nil
}
