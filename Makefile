# general variables
BUILD_ROOT=${PWD}
GIT_COMMIT=$(shell cd ${BUILD_ROOT} && git rev-list -n 1 --abbrev-commit HEAD)

# distribution variables
DIST=${BUILD_ROOT}/dist
DIST_FILE=dist-${GIT_COMMIT}.tar.gz

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
STATIC_SOURCE_FILES=$(shell find ${STATIC_ROOT} | egrep '\.(html|css|js)$$')

build: ${DIST} server client

clean:
	rm -r ${DIST}

.PHONY : clean

dist: clean ${DIST_FILE}

server: ${MTRACK_BIN} ${STATIC_ROOT_DIST}

client: ${MTRACK_CLIENT_BIN}

${DIST_FILE}: build
	tar cvzf $@ ${DIST}


# NOTE
# the symbolic links made for ${MTRACK_BIN} and ${MTRACK_CLIENT_BIN}
# may not be of much use in a production scenario. they are provided
# for more rapid development.

${MTRACK_BIN}: ${MTRACK_VERSION_BIN}
	rm -f $@
	ln -s ${MTRACK_VERSION_BIN} $@

${MTRACK_CLIENT_BIN}: ${MTRACK_CLIENT_VERSION_BIN}
	rm -f $@
	ln -s ${MTRACK_CLIENT_VERSION_BIN} $@

${MTRACK_VERSION_BIN}: ${MTRACK_SRC_FILES}
	cd ${BUILD_ROOT} && go build -o $@

${MTRACK_CLIENT_VERSION_BIN}: ${MTRACK_CLIENT_SRC_FILES}
	cd ${BUILD_ROOT} && go build -o $@ ./${MTRACK_CLIENT_REL_PATH}

${DIST}:
	mkdir -p $@

${STATIC_ROOT_DIST}: ${STATIC_SOURCE_FILES}
	echo ${STATIC_SOURCE_FILES}
	mkdir -p $@
	rsync -auv ${STATIC_ROOT}/ $@
