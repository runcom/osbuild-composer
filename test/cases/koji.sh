#!/bin/bash
set -euo pipefail

OSBUILD_COMPOSER_TEST_DATA=/usr/share/tests/osbuild-composer/

# Get OS data.
source /etc/os-release
ARCH=$(uname -m)

# Colorful output.
function greenprint {
    echo -e "\033[1;32m${1}\033[0m"
}

# Provision the software under tet.
/usr/libexec/osbuild-composer-test/provision.sh

greenprint "Defining distro selector"
# if the distro is RHEL 8.4 the distro includes the minor release number
if [[ "${ID}-${VERSION_ID}" == "rhel-8.4" ]]; then
    DISTRO_SELECTOR="${ID}-${VERSION_ID//.}"
# otherwise the minor release number can be dropped
else
    DISTRO_SELECTOR="${ID}-${VERSION_ID%.*}"
fi

greenprint "Adding podman dnsname plugin"
if [[ $ID == rhel || $ID == centos ]]; then
  sudo cp /usr/share/tests/osbuild-composer/vendor/87-podman-bridge.conflist /etc/cni/net.d/
  sudo cp /usr/share/tests/osbuild-composer/vendor/dnsname /usr/libexec/cni/
fi

greenprint "Starting containers"
sudo /usr/libexec/osbuild-composer-test/run-koji-container.sh start

greenprint "Copying custom worker config"
sudo mkdir -p /etc/osbuild-worker
sudo cp "${OSBUILD_COMPOSER_TEST_DATA}"/composer/osbuild-worker.toml \
    /etc/osbuild-worker/

greenprint "Adding kerberos config"
sudo cp \
    /tmp/osbuild-composer-koji-test/client.keytab \
    /etc/osbuild-composer/client.keytab
sudo cp \
    /tmp/osbuild-composer-koji-test/client.keytab \
    /etc/osbuild-worker/client.keytab
sudo cp \
    "${OSBUILD_COMPOSER_TEST_DATA}"/kerberos/krb5-local.conf \
    /etc/krb5.conf.d/local

greenprint "Adding the testsuite's CA cert to the system trust store"
sudo cp \
    /etc/osbuild-composer/ca-crt.pem \
    /etc/pki/ca-trust/source/anchors/osbuild-composer-tests-ca-crt.pem
sudo update-ca-trust

greenprint "Restarting composer to pick up new config"
sudo systemctl restart osbuild-composer
sudo systemctl restart osbuild-worker\@1

greenprint "Testing Koji"
koji --server=http://localhost:8080/kojihub --user=osbuild --password=osbuildpass --authtype=password hello

greenprint "Creating Koji task"
koji --server=http://localhost:8080/kojihub --user kojiadmin --password kojipass --authtype=password make-task image

greenprint "Pushing compose to Koji"
sudo /usr/libexec/osbuild-composer-test/koji-compose.py "$DISTRO_SELECTOR" "${ARCH}"

greenprint "Show Koji task"
koji --server=http://localhost:8080/kojihub taskinfo 1
koji --server=http://localhost:8080/kojihub buildinfo 1

greenprint "Run the integration test"
sudo /usr/libexec/osbuild-composer-test/osbuild-koji-tests

greenprint "Stopping containers"
sudo /usr/libexec/osbuild-composer-test/run-koji-container.sh stop

greenprint "Removing generated CA cert"
sudo rm \
    /etc/pki/ca-trust/source/anchors/osbuild-composer-tests-ca-crt.pem
sudo update-ca-trust
