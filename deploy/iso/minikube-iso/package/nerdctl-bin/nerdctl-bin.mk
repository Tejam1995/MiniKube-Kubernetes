################################################################################
#
# nerdctl-bin
#
################################################################################

NERDCTL_BIN_VERSION = v0.17.1
NERDCTL_BIN_COMMIT = 20233c26d26f11ca73a9a775fba87ca884ab14d2
NERDCTL_BIN_SITE = https://github.com/containerd/nerdctl/releases/download/$(NERDCTL_BIN_VERSION)
NERDCTL_BIN_SOURCE = nerdctl-$(subst v,,$(NERDCTL_BIN_VERSION))-linux-amd64.tar.gz
NERDCTL_BIN_STRIP_COMPONENTS = 0

define NERDCTL_BIN_INSTALL_TARGET_CMDS
	$(INSTALL) -D -m 0755 \
		$(@D)/nerdctl \
		$(TARGET_DIR)/usr/bin
endef

$(eval $(generic-package))
