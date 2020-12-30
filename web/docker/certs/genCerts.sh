#!/usr/bin/env bash

NAME=ebidlocal.local # Use your own domain name
# Generate a private key
openssl genrsa -out $NAME.key 2048
# Create a certificate-signing request
openssl req -new -key $NAME.key -out $NAME.csr
# Create a config file for the extensions
>$NAME.ext cat <<-EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = $NAME # Be sure to include the domain name here because Common Name is not so commonly honoured by itself
#DNS.2 = bar.$NAME # Optionally, add additional domains (I've added a subdomain here)
#IP.1 = 192.168.0.13 # Optionally, add an IP address (if the connection which you have planned requires it)
EOF
# Create the signed certificate
openssl x509 -req -in $NAME.csr -CA myCA.pem -CAkey myCA.key -CAcreateserial \
-out $NAME.crt -days 825 -sha256 -extfile $NAME.ext

# 1. Import myCA.pem as an Authority in your Chrome settings (Settings > Manage certificates > Authorities > Import)
# 2. Use the $NAME.crt and $NAME.key files in your server

# Extra steps (for Mac, at least):
# 1. Import the CA cert at "File > Import file", then also find it in the list, right click it, expand "> Trust", and select "Always"
# 2. Add extendedKeyUsage=serverAuth,clientAuth below basicConstraints=CA:FALSE, and make sure you set the "CommonName" to the same as $NAME when it's asking for setup
# 3. You can check your work

# openssl verify -CAfile myCA.pem -verify_hostname bar.mydomain.com mydomain.com.crt

# References
# * https://stackoverflow.com/questions/7580508/getting-chrome-to-accept-self-signed-localhost-certificate
# * https://www.digitalocean.com/community/tutorials/how-to-create-a-self-signed-ssl-certificate-for-nginx-on-centos-7
