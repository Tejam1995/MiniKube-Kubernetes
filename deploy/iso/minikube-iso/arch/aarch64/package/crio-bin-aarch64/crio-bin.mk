################################################################################
#
# cri-o
#
################################################################################

CRIO_BIN_AARCH64_VERSION = v1.22.3
CRIO_BIN_AARCH64_COMMIT = d93b2dfb8d0f2ad0f8b9061d941e3b216baa5814
CRIO_BIN_AARCH64_SITE = https://github.com/cri-o/cri-o/archive
CRIO_BIN_AARCH64_SOURCE = $(CRIO_BIN_AARCH64_VERSION).tar.gz
CRIO_BIN_AARCH64_DEPENDENCIES = host-go libgpgme
CRIO_BIN_AARCH64_GOPATH = $(@D)/_output
CRIO_BIN_AARCH64_ENV = \
	$(GO_TARGET_ENV) \
	CGO_ENABLED=1 \
	GO111MODULE=off \
	GOPATH="$(CRIO_BIN_AARCH64_GOPATH)" \
	PATH=$(CRIO_BIN_AARCH64_GOPATH)/bin:$(BR_PATH) \
	GOARCH=arm64


define CRIO_BIN_AARCH64_USERS
	- -1 crio-admin -1 - - - - -
	- -1 crio       -1 - - - - -
endef

define CRIO_BIN_AARCH64_CONFIGURE_CMDS
	mkdir -p $(CRIO_BIN_AARCH64_GOPATH)/src/github.com/cri-o
	ln -sf $(@D) $(CRIO_BIN_AARCH64_GOPATH)/src/github.com/cri-o/cri-o
	# disable the "automatic" go module detection
	sed -e 's/go help mod/false/' -i $(@D)/Makefile
endef

define CRIO_BIN_AARCH64_BUILD_CMDS
	mkdir -p $(@D)/bin
	$(CRIO_BIN_AARCH64_ENV) $(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) COMMIT_NO=$(CRIO_BIN_AARCH64_COMMIT) PREFIX=/usr binaries
endef

define CRIO_BIN_AARCH64_INSTALL_TARGET_CMDS
	mkdir -p $(TARGET_DIR)/usr/share/containers/oci/hooks.d
	mkdir -p $(TARGET_DIR)/etc/containers/oci/hooks.d
	mkdir -p $(TARGET_DIR)/etc/crio/crio.conf.d

	$(INSTALL) -Dm755 \
		$(@D)/bin/crio \
		$(TARGET_DIR)/usr/bin/crio
	$(INSTALL) -Dm755 \
		$(@D)/bin/pinns \
		$(TARGET_DIR)/usr/bin/pinns
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_AARCH64_PKGDIR)/crio.conf \
		$(TARGET_DIR)/etc/crio/crio.conf
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_AARCH64_PKGDIR)/policy.json \
		$(TARGET_DIR)/etc/containers/policy.json
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_AARCH64_PKGDIR)/registries.conf \
		$(TARGET_DIR)/etc/containers/registries.conf
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_AARCH64_PKGDIR)/02-crio.conf \
		$(TARGET_DIR)/etc/crio/crio.conf.d/02-crio.conf

	mkdir -p $(TARGET_DIR)/etc/sysconfig
	echo 'CRIO_OPTIONS="--log-level=debug"' > $(TARGET_DIR)/etc/sysconfig/crio
endef

define CRIO_BIN_AARCH64_INSTALL_INIT_SYSTEMD
	$(MAKE) $(TARGET_CONFIGURE_OPTS) -C $(@D) install.systemd DESTDIR=$(TARGET_DIR) PREFIX=$(TARGET_DIR)/usr
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_AARCH64_PKGDIR)/crio.service \
		$(TARGET_DIR)/usr/lib/systemd/system/crio.service
	$(INSTALL) -Dm644 \
		$(CRIO_BIN_AARCH64_PKGDIR)/crio-wipe.service \
		$(TARGET_DIR)/usr/lib/systemd/system/crio-wipe.service
	$(call link-service,crio.service)
	$(call link-service,crio-shutdown.service)
endef

$(eval $(generic-package))
