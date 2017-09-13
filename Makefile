BEAT_NAME=dbeat
BEAT_PATH=github.com/Axway/elasticsearch-docker-beat
BEAT_GOPATH=$(firstword $(subst :, ,${GOPATH}))
BEAT_URL=https://${BEAT_PATH}
SYSTEM_TESTS=false
TEST_ENVIRONMENT=false
ES_BEATS?=./vendor/github.com/elastic/beats
GOPACKAGES=$(shell glide novendor)
PREFIX?=.
NOTICE_FILE=NOTICE

# Path to the libbeat Makefile
-include $(ES_BEATS)/libbeat/scripts/Makefile

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
	rm -f elasticsearch-docker-beat
	docker build -t axway/elasticsearch-docker-beat:latest .

.PHONY: push-image
push-image: create-image
	docker save axway/elasticsearch-docker-beat -o dbeat.dimg
	docker-machine scp dbeat.dimg default:/tmp/
	docker-machine ssh default docker load -i /tmp/dbeat.dimg
	rm dbeat.dimg
	docker-machine ssh default rm /tmp/dbeat.dimg

.PHONY: create-image-test
create-image-test:
	rm -f elasticsearch-docker-beat
	docker build -t axway/elasticsearch-docker-beat:test .

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
