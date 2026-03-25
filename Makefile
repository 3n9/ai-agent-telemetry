GOPATH_BIN := $(shell /opt/homebrew/bin/go env GOPATH)/bin
GO         := /opt/homebrew/bin/go
LDFLAGS    := -s -w
DIST       := dist

CMDS := ai-log ai-log-report

INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: all clean build install uninstall install-global $(CMDS)

all: \
	$(DIST)/darwin-arm64/ai-log \
	$(DIST)/darwin-arm64/ai-log-report \
	$(DIST)/darwin-amd64/ai-log \
	$(DIST)/darwin-amd64/ai-log-report \
	$(DIST)/linux-amd64/ai-log \
	$(DIST)/linux-amd64/ai-log-report \
	$(DIST)/windows-amd64/ai-log.exe \
	$(DIST)/windows-amd64/ai-log-report.exe

# ── macOS arm64 (Apple Silicon) ───────────────────────────────────────────────
$(DIST)/darwin-arm64/%: GOARCH=arm64
$(DIST)/darwin-arm64/%: GOOS=darwin
$(DIST)/darwin-arm64/ai-log:
	@mkdir -p $(DIST)/darwin-arm64
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $@ ./cmd/ai-log

$(DIST)/darwin-arm64/ai-log-report:
	@mkdir -p $(DIST)/darwin-arm64
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $@ ./cmd/ai-log-report

# ── macOS amd64 (Intel) ───────────────────────────────────────────────────────
$(DIST)/darwin-amd64/ai-log:
	@mkdir -p $(DIST)/darwin-amd64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $@ ./cmd/ai-log

$(DIST)/darwin-amd64/ai-log-report:
	@mkdir -p $(DIST)/darwin-amd64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $@ ./cmd/ai-log-report

# ── Linux amd64 ───────────────────────────────────────────────────────────────
$(DIST)/linux-amd64/ai-log:
	@mkdir -p $(DIST)/linux-amd64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $@ ./cmd/ai-log

$(DIST)/linux-amd64/ai-log-report:
	@mkdir -p $(DIST)/linux-amd64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $@ ./cmd/ai-log-report

# ── Windows amd64 ─────────────────────────────────────────────────────────────
$(DIST)/windows-amd64/ai-log.exe:
	@mkdir -p $(DIST)/windows-amd64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $@ ./cmd/ai-log

$(DIST)/windows-amd64/ai-log-report.exe:
	@mkdir -p $(DIST)/windows-amd64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $@ ./cmd/ai-log-report

# ── dev shortcut (native) ─────────────────────────────────────────────────────
build:
	$(GO) build -o $(DIST)/ai-log ./cmd/ai-log
	$(GO) build -o $(DIST)/ai-log-report ./cmd/ai-log-report

clean:
	rm -rf $(DIST)

# ── install (native build → ~/.local/bin) ────────────────────────────────────
install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(DIST)/ai-log $(DIST)/ai-log-report $(INSTALL_DIR)/
	@echo "installed to $(INSTALL_DIR)"
	@echo "make sure $(INSTALL_DIR) is in your PATH"

# -- uninstall from ~/.local/bin ─────────────────────────────────────────────────
uninstall:
	rm -f $(INSTALL_DIR)/ai-log $(INSTALL_DIR)/ai-log-report
	@echo "uninstalled from $(INSTALL_DIR)"

# ── global installation for AI agents ─────────────────────────────────────────
install-global:
	@bash scripts/install-global.sh

