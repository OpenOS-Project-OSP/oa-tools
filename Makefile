# artisan/Makefile
# Fallback se non viene passata da coa
VERSION ?= 0.0.0-dev

# Directories
OA_DIR = oa
COA_DIR = coa

# Binaries
OA_BIN = $(OA_DIR)/oa
COA_BIN = $(COA_DIR)/coa

all: build_oa build_coa
	@echo "--------------------------------------"
	@echo "Hatching completed successfully! 🐣"
	@echo "coa Brain (Go):   ./$(COA_BIN)"
	@echo "oa Workhorse (C): ./$(OA_BIN)"
	@echo "--------------------------------------"

build_oa:
	@echo "  MAKING oa..."
	@$(MAKE) -C $(OA_DIR) VERSION="$(VERSION)"

build_coa:
	@echo "  MAKING coa..."
	@cd $(COA_DIR) && go build -o coa ./src

clean:
	@echo "  Pulizia in corso..."
	@$(MAKE) -C $(OA_DIR) clean
	@rm -f $(COA_BIN)
	@rm -f $(COA_DIR)/plan_coa_tmp.json

.PHONY: all build_oa build_coa clean
