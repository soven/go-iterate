GOCMD := "go"
GOBIN := $(shell command -v $(GOCMD) 2> /dev/null)
CURDIR := $(shell pwd)
GEN_ITER_KIT_CMD := $(CURDIR)/cmd/gen-go-iter-kit/main.go

iter:
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
.PHONY: iter

clean:
	rm -f gen_*

pre_gen_all: clean
	while IFS=, read -r typ title zero pkg; do \
  		if [ -z "$$title" ]; then \
  			$(GOBIN) run $(GEN_ITER_KIT_CMD) -target $$typ -zero $$zero -package $$pkg; \
  		else \
  			$(GOBIN) run $(GEN_ITER_KIT_CMD) -target $$typ -prefix $$title -zero $$zero -package $$pkg; \
  		fi; \
  	done < pre_gen_context.csv
.PHONY: pre_gen_all
