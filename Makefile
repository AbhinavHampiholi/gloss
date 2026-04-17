# Build and install git-gloss.
#
# Default install target is ~/.local/bin (no sudo needed, on PATH by
# default on most modern setups). Override with `make install PREFIX=/usr/local`
# for a system-wide install.

BIN      := git-gloss
PREFIX   ?= $(HOME)/.local
BINDIR   := $(PREFIX)/bin

.PHONY: all build install uninstall run clean check

all: build

build:
	go build -o $(BIN) .

install: build
	@mkdir -p $(BINDIR)
	@install -m 0755 $(BIN) $(BINDIR)/$(BIN)
	@echo "installed $(BINDIR)/$(BIN)"
	@case ":$$PATH:" in \
	  *":$(BINDIR):"*) ;; \
	  *) echo; \
	     echo "note: $(BINDIR) is not on your PATH."; \
	     echo "  add to ~/.zshrc or ~/.bashrc:"; \
	     echo "    export PATH=\"$(BINDIR):\$$PATH\"" ;; \
	esac

uninstall:
	@rm -f $(BINDIR)/$(BIN)
	@echo "removed $(BINDIR)/$(BIN)"

run: build
	./$(BIN) $(ARGS)

check:
	go vet ./...
	go build ./...

clean:
	rm -f $(BIN)
