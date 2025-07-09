SERVER_EXECUTABLE=pocketproxyserver
KOBO_MOD_DIR=${PWD}/kobo-mod
KOBO_MAKE=podman run --volume="${KOBO_MOD_DIR}:${KOBO_MOD_DIR}" --workdir="${KOBO_MOD_DIR}" --env=HOME --entrypoint="make" --rm -it ghcr.io/pgaskin/nickeltc:1.0


all:
	mkdir -p bin/proxy-server
	mkdir -p bin/kobo-mod
	cd proxy-server && make build
	${KOBO_MAKE} \
		&& ${KOBO_MAKE} koboroot \
		&& mv -f ${KOBO_MOD_DIR}/libpocketproxy.so bin/kobo-mod/ \
		&& mv -f ${KOBO_MOD_DIR}/KoboRoot.tgz bin/kobo-mod/		

clean:
	cd proxy-server && make clean
	rm -f ../bin/kobo-mod/libpocketproxy
	rm -f ../bin/kobo-mod/KoboRoot.tgz
	rm -f ${KOBO_MOD_DIR}/src/pocketproxy.o
