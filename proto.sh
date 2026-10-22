#!/usr/bin/env sh
# genera el codigo go de los .proto, osea los mensajes y el grpc
# todo corre dentro de docker y solo necesitas docker prendido
set -e

docker build -t soc-protoc tools/protoc

# pwd -W es pa git bash en windows y si no existe usa pwd normal
ROOT="$(pwd -W 2>/dev/null || pwd)"

docker run --rm -v "$ROOT:/work" -w //work soc-protoc sh -c '
# gen <modulo> <servicio> -> saca el codigo de un proto y lo deja en <modulo>/gen/<servicio>pb/
# cada modulo guarda su propia copia pq cada servicio es un modulo go aparte
gen() {
  mkdir -p "$1/gen/$2pb"
  protoc --proto_path=proto \
    --go_out="$1/gen/$2pb" --go_opt=paths=source_relative \
    --go-grpc_out="$1/gen/$2pb" --go-grpc_opt=paths=source_relative \
    "proto/$2/v1/$2.proto"
}

# quien necesita que
gen backend accounts     # accounts es servidor de accounts
gen backend events       # y cliente de events pa avisar los logins
gen events events        # events es servidor de events
gen status status        # status es servidor de status
gen middleware accounts  # middleware es cliente y servidor de los tres
gen middleware events    # porque es el proxy
gen middleware status
'

echo "listo, codigo generado en */gen/"
