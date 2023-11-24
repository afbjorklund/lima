package hostagent

import (
	"context"

	"github.com/lima-vm/lima/pkg/windows"
	"github.com/lima-vm/sshocker/pkg/ssh"
)

func forwardTCP(ctx context.Context, sshConfig *ssh.SSHConfig, addr string, port int, local, remote, verb string) error {
	return forwardSSH(ctx, sshConfig, addr, port, local, remote, verb, false)
}

func getFreeVSockPort() (int, error) {
	return windows.GetRandomFreeVSockPort(0, 2147483647)
}
