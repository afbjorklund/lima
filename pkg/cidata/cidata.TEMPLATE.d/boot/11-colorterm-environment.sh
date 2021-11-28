#!/bin/sh
set -eux

! grep -q "COLORTERM" /etc/ssh/sshd_config || exit 0

# accept any incoming COLORTERM environment variable
sed -i 's/^AcceptEnv LANG LC_\*$/AcceptEnv COLORTERM TERMTHEME LANG LC_*/' /etc/ssh/sshd_config
! test -f /sbin/openrc-init || rc-service --ifstarted sshd reload
! command -v systemctl >/dev/null 2>&1 || systemctl reload ssh\*
