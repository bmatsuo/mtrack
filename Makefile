# general variables
BUILD_ROOT=${PWD}
GIT_COMMIT=$(shell cd ${BUILD_ROOT} && git rev-list -n 1 --abbrev-commit HEAD)
SQLITE3_DB_PATH=${PWD}/data/mtrack.sqlite

# distribution variables
DIST=${BUILD_ROOT}/dist
DIST_FILE=${BUILD_ROOT}/dist-${GIT_COMMIT}.tar.gz

# server binary
MTRACK_SRC_FILES=$(shell find ${BUILD_ROOT} -name '*.go' | egrep -v '^${BUILD_ROOT}/tools')
MTRACK_BIN=${DIST}/mtrack
MTRACK_VERSION_BIN=${MTRACK_BIN}-${GIT_COMMIT}

# command-line client binary
MTRACK_CLIENT_REL_PATH=tools/mtrack-client
MTRACK_CLIENT_PATH=${BUILD_ROOT}/${MTRACK_CLIENT_REL_PATH}
MTRACK_CLIENT_SRC_FILES=$(shell find ${MTRACK_CLIENT_PATH} -name '*.go')
MTRACK_CLIENT_BIN=${DIST}/mtrack-client
MTRACK_CLIENT_VERSION_BIN=${MTRACK_CLIENT_BIN}-${GIT_COMMIT}

# static files (eventually some kind of asset pipeline)
STATIC_ROOT_REL=http/static
STATIC_ROOT=${BUILD_ROOT}/${STATIC_ROOT_REL}
STATIC_ROOT_DIST=${DIST}/static
STATIC_ROOT_VERSION_DIST=${STATIC_ROOT_DIST}-${GIT_COMMIT}
STATIC_SOURCE_FILES=$(shell find ${STATIC_ROOT} | egrep '\.(html|css|js)$$')

help:
	@echo "make [command]:" 1>&2
	@echo "\tbuild        compile both the server (including static assets). build client programs" 1>&2
	@echo "\tclean        remove the dist directory" 1>&2
	@echo "\tdist         create an archive file for distribution" 1>&2
	@echo "\tclient       compile just the client program" 1>&2
	@echo "\tserver       compile just the server program" 1>&2
	@echo "\tstart        start a server that laods static files from the development directory" 1>&2
	@echo "\tstart-dist   start a server that uses a fixed set of static files." 1>&2
	@exit 1

.PHONY : help

start: ${MTRACK_VERSION_BIN}
	${MTRACK_VERSION_BIN} -media=./data/media

start-dist: ${MTRACK_VERSION_BIN} ${STATIC_ROOT_VERSION_DIST}
	${MTRACK_VERSION_BIN} -media=./data/media -http.static='${STATIC_ROOT_VERSION_DIST}'

build: ${DIST} server client

clean:
	[ -d ${DIST} ] && rm -r ${DIST} || echo -n

.PHONY : clean

dist: clean ${DIST_FILE}

server: ${MTRACK_BIN} ${STATIC_ROOT_DIST}

client: ${MTRACK_CLIENT_BIN}

drop:
	rm ${SQLITE3_DB_PATH}

.PHONY : drop

${DIST_FILE}: ${DIST} ${MTRACK_VERSION_BIN} ${STATIC_ROOT_VERSION_DIST}
	cd $(shell dirname ${DIST}) && tar cvzf $@ $(shell basename ${DIST})

# NOTE
# the symbolic links made for ${MTRACK_BIN}, ${STATIC_ROOT_VERSION_DIST},
# and ${MTRACK_CLIENT_BIN} may not be of much use in a production scenario.
# they are provided for more rapid development.

${MTRACK_BIN}: ${MTRACK_VERSION_BIN}
	rm -f $@
	ln -s ${MTRACK_VERSION_BIN} $@

${MTRACK_CLIENT_BIN}: ${MTRACK_CLIENT_VERSION_BIN}
	rm -f $@
	ln -s ${MTRACK_CLIENT_VERSION_BIN} $@

${STATIC_ROOT_DIST}: ${STATIC_ROOT_VERSION_DIST}
	rm -f $@
	ln -s ${STATIC_ROOT_VERSION_DIST} $@

${MTRACK_VERSION_BIN}: ${MTRACK_SRC_FILES}
	go get -d .
	go build -o $@

${MTRACK_CLIENT_VERSION_BIN}: ${MTRACK_CLIENT_SRC_FILES}
	go get -d ./${MTRACK_CLIENT_REL_PATH}
	go build -o $@ ./${MTRACK_CLIENT_REL_PATH}

${DIST}:
	mkdir -p $@

${STATIC_ROOT_VERSION_DIST}: ${STATIC_SOURCE_FILES}
	mkdir -p $@
	rsync -auv ${STATIC_ROOT}/ $@
