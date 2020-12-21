GOCMD := "go"
GOBIN := $(shell command -v $(GOCMD) 2> /dev/null)
GEN_ITER_KIT_CMD_PATH := cmd/gen-go-iter-kit
CURDIR := $(shell pwd)
VERSION := v0.0.1-beta
MOCKERYCMD := "mockery"
MOCKERYBIN := $(shell command -v $(MOCKERYCMD) 2> /dev/null)
MOCK_SOURCE_DIR := $(CURDIR)/internal/testground/resembled
MOCKS_DIR := $(MOCK_SOURCE_DIR)/mocks
INSTALL_GEN_ITER_KIT_CMD := $(CURDIR)/$(GEN_ITER_KIT_CMD_PATH)
GEN_ITER_KIT_CMD := $(CURDIR)/$(GEN_ITER_KIT_CMD_PATH)/main.go
ASSEMBLE_TEMPLATES_CMD := $(CURDIR)/cmd/assemble-templates/main.go
ASSEMBLE_TEMPLATES_PATH := templates
ASSEMBLE_TEMPLATES_TYPE_FROM := Type
ASSEMBLE_TEMPLATES_PACKAGE_FROM := resembled
ASSEMBLE_TEMPLATES_ZERO_FROM := Zero
ASSEMBLE_TEMPLATES_PREFIX_FROM := Prefix

install:
	$(GOBIN) install -ldflags="-X 'main.Version=$(VERSION)'" "$(INSTALL_GEN_ITER_KIT_CMD)"

kit:
ifdef GOBIN
ifdef PATH
ifdef TYPE
ifdef PACKAGE
ifdef ZERO
ifdef PREFIX
	$(GOBIN) run $(GEN_ITER_KIT_CMD) \
	-target $(TYPE) -package $(PACKAGE) -prefix $(PREFIX) -path $(PATH) -zero $(ZERO)
else
	$(GOBIN) run $(GEN_ITER_KIT_CMD) \
	-target $(TYPE) -package $(PACKAGE) -path $(PATH) -zero $(ZERO)
endif
else
	@echo "define ZERO"
endif
else
	@echo "define PACKAGE"
endif
else
	@echo "define TYPE"
endif
else
	@echo "define PATH"
endif
else
	@echo "$(GOCMD) not found in PATH"
endif
.PHONY: kit

clean:
	rm -f gen_*

gen_kit_all: clean
	while IFS=, read -r typ title zero pkg; do \
  		if [ -z "$$title" ]; then \
  			$(GOBIN) run $(GEN_ITER_KIT_CMD) -target $$typ -zero $$zero -package $$pkg; \
  		else \
  			$(GOBIN) run $(GEN_ITER_KIT_CMD) -target $$typ -prefix $$title -zero $$zero -package $$pkg; \
  		fi; \
  	done < pre_gen_context.csv
.PHONY: gen_kit_all

clean_assembled:
	rm -rf $(ASSEMBLE_TEMPLATES_PATH) && mkdir -p $(ASSEMBLE_TEMPLATES_PATH)

assemble_all: clean_assembled
	$(GOBIN) run $(ASSEMBLE_TEMPLATES_CMD) -path $(ASSEMBLE_TEMPLATES_PATH) \
		-target $(ASSEMBLE_TEMPLATES_TYPE_FROM) -prefix $(ASSEMBLE_TEMPLATES_PREFIX_FROM) \
		-zero $(ASSEMBLE_TEMPLATES_ZERO_FROM) -package $(ASSEMBLE_TEMPLATES_PACKAGE_FROM)
.PHONY: assemble_all

clean_mocks:
	rm -rf $(MOCKS_DIR)

mocks: clean_mocks
	mockery --dir=$(MOCK_SOURCE_DIR) --output=$(MOCKS_DIR) --name=PrefixIterator
	mockery --dir=$(MOCK_SOURCE_DIR) --output=$(MOCKS_DIR) --name=PrefixChecker
	mockery --dir=$(MOCK_SOURCE_DIR) --output=$(MOCKS_DIR) --name=PrefixEnumChecker
	mockery --dir=$(MOCK_SOURCE_DIR) --output=$(MOCKS_DIR) --name=PrefixConverter
	mockery --dir=$(MOCK_SOURCE_DIR) --output=$(MOCKS_DIR) --name=PrefixEnumConverter
	mockery --dir=$(MOCK_SOURCE_DIR) --output=$(MOCKS_DIR) --name=PrefixHandler
	mockery --dir=$(MOCK_SOURCE_DIR) --output=$(MOCKS_DIR) --name=PrefixEnumHandler