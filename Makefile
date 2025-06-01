SERVER_EXECUTABLE=pocketproxyserver
KOBO_MOD_DIR=${PWD}/kobo-mod
KOBO_MAKE=docker run --volume="${KOBO_MOD_DIR}:${KOBO_MOD_DIR}" --workdir="${KOBO_MOD_DIR}" --env=HOME --entrypoint="make" --rm -it ghcr.io/pgaskin/nickeltc:1.0


all:
	cd proxy-server && make build
	sudo ${KOBO_MAKE} \
		&& sudo ${KOBO_MAKE} koboroot \
		&& mv ${KOBO_MOD_DIR}/libpocketproxy.so bin \
		&& mv ${KOBO_MOD_DIR}/KoboRoot.tgz bin

clean:
	cd proxy-server && make clean
	rm -f ../bin/libpocketproxy
	rm -f ../bin/KoboRoot.tgz
