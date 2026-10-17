#!/bin/sh
# Creates, once, a self-signed code signing identity in the login keychain,
# which dev and build sign Elir with. macOS keeps the Accessibility, Screen
# Recording and Microphone permissions of an app by its signature: with an
# ad-hoc one every rebuild is a new app to it and asks again; with this
# identity rebuilds keep them.
set -eu

name="${ELIR_SIGNING_IDENTITY:-Kit Local Signing}"
keychain="$HOME/Library/Keychains/login.keychain-db"

if security find-certificate -c "$name" "$keychain" >/dev/null 2>&1; then
  echo "\"$name\" is in the login keychain already."
  exit 0
fi

dir=$(mktemp -d)
trap 'rm -rf "$dir"' EXIT
cat >"$dir/cert.cnf" <<CNF
[req]
distinguished_name = dn
x509_extensions = ext
prompt = no
[dn]
CN = $name
[ext]
basicConstraints = critical, CA:false
keyUsage = critical, digitalSignature
extendedKeyUsage = critical, codeSigning
CNF
openssl req -x509 -newkey rsa:2048 -nodes -days 3650 -config "$dir/cert.cnf" \
  -keyout "$dir/key.pem" -out "$dir/cert.pem" 2>/dev/null
# The algorithms the Keychain imports, in OpenSSL 3 and LibreSSL alike.
openssl pkcs12 -export -inkey "$dir/key.pem" -in "$dir/cert.pem" -name "$name" \
  -keypbe PBE-SHA1-3DES -certpbe PBE-SHA1-3DES -macalg sha1 \
  -out "$dir/identity.p12" -passout pass:kit
# -T lets codesign use the key without asking.
security import "$dir/identity.p12" -k "$keychain" -P kit -T /usr/bin/codesign >/dev/null
echo "Created \"$name\" in the login keychain."
