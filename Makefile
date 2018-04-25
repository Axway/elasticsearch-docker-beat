BEAT_NAME=elasticsearch-docker-beat
BEAT_PATH=github.com/Axway/$(BEAT_NAME)
BEAT_GOPATH=$(firstword $(subst :, ,${GOPATH}))
BEAT_URL=https://${BEAT_PATH}
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS?=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
PREFIX?=.
NOTICE_FILE=NOTICE

BUILD := $(shell git rev-parse HEAD | cut -c1-8)
LDFLAGS := -s -w

# target from the libbeat Makefile, with the ldflags added
# -include $(ES_BEATS)/libbeat/scripts/Makefile
GOFILES = $(shell find beater -type f -name '*.go')
GOFILES_ALL = $(GOFILES) $(shell find $(ES_BEATS) -type f -name '*.go')
$(BEAT_NAME): $(GOFILES_ALL) ## @build build the beat application
	go build -ldflags "-X=$(BEAT_PATH)/beater.Build=$(BUILD) $(LDFLAGS)"

# Initial beat setup
.PHONY: setup
setup: copy-vendor
	make update

# Copy beats into vendor directory
.PHONY: copy-vendor
copy-vendor:
	mkdir -p vendor/github.com/elastic/
	cp -R ${BEAT_GOPATH}/src/github.com/elastic/beats vendor/github.com/elastic/
	rm -rf vendor/github.com/elastic/beats/.git

.PHONY: create-image
create-image:
	rm -f $(BEAT_NAME)
	docker build -t axway/$(BEAT_NAME):latest .

.PHONY: push-image
push-image: create-image
	docker save axway/$(BEAT_NAME) -o dbeat.dimg
	docker-machine scp dbeat.dimg default:/tmp/
	docker-machine ssh default docker load -i /tmp/dbeat.dimg
	rm dbeat.dimg
	docker-machine ssh default rm /tmp/dbeat.dimg

.PHONY: create-image-test
create-image-test:
	rm -f $(BEAT_NAME)
	docker build -t axway/$(BEAT_NAME):test .

.PHONY: update-deps
update-deps:


.PHONY: git-init
git-init:
	git init
	git add README.md CONTRIBUTING.md
	git commit -m "Initial commit"
	git add LICENSE
	git commit -m "Add the LICENSE"
	git add .gitignore
	git commit -m "Add git settings"
	git add .
	git reset -- .travis.yml
	git commit -m "Add dbeat"
	git add .travis.yml
	git commit -m "Add Travis CI"

# This is called by the beats packer before building starts
.PHONY: before-build
before-build:

# Collects all dependencies and then calls update
.PHONY: collect
collect:
