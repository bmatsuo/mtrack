# general variables
BUILD_ROOT=${PWD}
GIT_COMMIT=$(shell cd ${BUILD_ROOT} && git rev-list -n 1 --abbrev-commit HEAD)
SQLITE3_DB_PATH=${PWD}/data/mtrack.sqlite

# distribution variables
DIST_ROOT=${BUILD_ROOT}/dist
DIST=${DIST_ROOT}/mtrack-${GIT_COMMIT}
DIST_FILE=${DIST}.tar.gz
DIST_TOOLS=${BUILD_ROOT}/dist/mtrack-tools-${GIT_COMMIT}
DIST_TOOLS_FILE=${DIST_TOOLS}.tar.gz
DIST_ALL=${BUILD_ROOT}/dist/mtrack-all-${GIT_COMMIT}
DIST_ALL_FILE=${DIST_ALL}.tar.gz

# server binary
MTRACK_SRC_FILES=$(shell find ${BUILD_ROOT} -name '*.go' | egrep -v '^${BUILD_ROOT}/tools')
MTRACK_BIN=${DIST}/mtrack

# command-line client binary
MTRACK_CLIENT_REL_PATH=tools/mtrack-client
MTRACK_CLIENT_PATH=${BUILD_ROOT}/${MTRACK_CLIENT_REL_PATH}
MTRACK_CLIENT_SRC_FILES=$(shell find ${MTRACK_CLIENT_PATH} -name '*.go')
MTRACK_CLIENT_BIN=${DIST}/mtrack-client

# static files (eventually some kind of asset pipeline)
STATIC_ROOT_REL=http/static
STATIC_ROOT=${BUILD_ROOT}/${STATIC_ROOT_REL}
STATIC_ROOT_DIST=${DIST}/static
STATIC_SOURCE_FILES=$(shell find ${STATIC_ROOT} | egrep '\.(html|css|js)$$')

help:
	@echo "make [command]:" 1>&2
	@echo "\tbuild        compile both the server (including static assets). build client programs" 1>&2
	@echo "\tclean        remove build output, but not archived distributions" 1>&2
	@echo "\tspotless     remove all contents of the dist/ directory" 1>&2
	@echo "\tdist         create an archive file for distribution of the server" 1>&2
	@echo "\tdist-tools   create an archive file for distribution of the tooling" 1>&2
	@echo "\tdist-all     create an archive file for distribution of the server and tooling" 1>&2
	@echo "\tserver       compile just the server program" 1>&2
	@echo "\tclient       compile just the client tool" 1>&2
	@echo "\ttools        build all available tooling" 1>&2
	@echo "\tstart        start a server that laods static files from the development directory" 1>&2
	@echo "\tstart-dist   start a server that uses a fixed set of static files" 1>&2
	@echo "\tdrop         delete the development sqlite3 database" 1>&2
	@exit 1

.PHONY : help

start: server
	${MTRACK_BIN} -config=${PWD}/example.mtrack.toml

start-dist: server
	${MTRACK_BIN} -media=./data/media -http.static='${STATIC_ROOT_DIST}'

build: ${DIST} server client

clean:
	[ -d ${DIST_ROOT} ] && ls ${DIST_ROOT} | egrep -v '\.tar\.gz$$' | sed 's:^:${DIST_ROOT}/:' | xargs rm -r || echo -n

spotless:
	[ -d ${DIST_ROOT} ] && rm -r ${DIST_ROOT}/*

.PHONY : clean spotless

${DIST}:
	mkdir -p $@

dist: clean ${DIST_FILE}

${DIST_FILE}: ${DIST} server
	cd $(shell dirname ${DIST}) && tar cvzf $@ $(shell basename ${DIST})

dist-tools: clean ${DIST_TOOLS_FILE}

${DIST_TOOLS_FILE}: ${DIST} client
	cd $(shell dirname ${DIST}) && tar cvzf $@ $(shell basename ${DIST})

dist-all: clean ${DIST_ALL_FILE}

${DIST_ALL_FILE}: ${DIST} server client
	cd $(shell dirname ${DIST}) && tar cvzf $@ $(shell basename ${DIST})

server: ${MTRACK_BIN} ${STATIC_ROOT_DIST}

${MTRACK_BIN}: ${MTRACK_SRC_FILES}
	go get -d .
	go build -o $@

${STATIC_ROOT_DIST}: ${STATIC_SOURCE_FILES}
	mkdir -p $@
	rsync -auv ${STATIC_ROOT}/ $@

tools: client

client: ${MTRACK_CLIENT_BIN}

${MTRACK_CLIENT_BIN}: ${MTRACK_CLIENT_SRC_FILES}
	go get -d ./${MTRACK_CLIENT_REL_PATH}
	go build -o $@ ./${MTRACK_CLIENT_REL_PATH}

drop:
	rm ${SQLITE3_DB_PATH}

.PHONY : drop
