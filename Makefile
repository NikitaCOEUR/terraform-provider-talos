TAG ?= $(shell git describe --tag --always --dirty)
ARTIFACTS ?= _out
TEST_TIMEOUT ?= 600s

ifneq ($(origin TESTS), undefined)
	RUNARGS = -run='$(TESTS)'
endif

ifneq ($(origin CI), undefined)
	RUNARGS += -parallel=3
	RUNARGS += -timeout=40m
	RUNARGS += -exec="sudo -E"
endif

.PHONY: generate
generate:
	go generate ./pkg/talos
	go generate

.PHONY: testacc
testacc:
	# TF_CLI_CONFIG_FILE is set here to avoid using the user's .terraformrc file. Ref: https://github.com/hashicorp/terraform-plugin-sdk/issues/1171
	TF_CLI_CONFIG_FILE="thisfiledoesnotexist" TF_ACC=1 go test -v -failfast -cover $(RUNARGS) ./...

# Debug tests with TF_LOG for verbose Terraform output
.PHONY: testacc-debug
testacc-debug:
	TF_LOG=DEBUG TF_CLI_CONFIG_FILE="thisfiledoesnotexist" TF_ACC=1 go test -v -failfast -cover $(RUNARGS) ./... 2>&1 | tee testacc-debug.log

# Run specific upgrade tests to debug staged_if_needing_reboot behavior
.PHONY: testacc-staged
testacc-staged:
	TF_LOG=DEBUG TF_CLI_CONFIG_FILE="thisfiledoesnotexist" TF_ACC=1 go test -v -failfast -cover -timeout=40m \
		-run='TestAccTalosMachineConfigurationApplyResourceAutoStaged|TestAccTalosMachineConfigurationApplyResourceUpgradeWithFix|TestAccTalosMachineConfigurationApplyResourceUpgradeWithBug' \
		$(RUNARGS) ./pkg/talos/ 2>&1 | tee testacc-staged.log

# Run AutoStaged test only with debug logging
.PHONY: testacc-autostaged-debug
testacc-autostaged-debug:
	TF_LOG=DEBUG TF_CLI_CONFIG_FILE="thisfiledoesnotexist" TF_ACC=1 go test -v -failfast -timeout=40m \
		-run='TestAccTalosMachineConfigurationApplyResourceAutoStaged$$' \
		$(RUNARGS) ./pkg/talos/ 2>&1 | tee testacc-autostaged.log

# Run UpgradeWithFix test only with debug logging
.PHONY: testacc-upgradefix-debug
testacc-upgradefix-debug:
	TF_LOG=DEBUG TF_CLI_CONFIG_FILE="thisfiledoesnotexist" TF_ACC=1 go test -v -failfast -timeout=40m \
		-run='TestAccTalosMachineConfigurationApplyResourceUpgradeWithFix$$' \
		$(RUNARGS) ./pkg/talos/ 2>&1 | tee testacc-upgradefix.log


.PHONY: check-dirty
check-dirty: generate ## Verifies that source tree is not dirty
	@if test -n "`git status --porcelain`"; then echo "Source tree is dirty"; git status; exit 1 ; fi

build-debug:
	go build -gcflags='all=-N -l'

install:
	go install .

release-notes:
	mkdir -p $(ARTIFACTS)
	@ARTIFACTS=$(ARTIFACTS) ./hack/release.sh $@ $(ARTIFACTS)/RELEASE_NOTES.md $(TAG)
