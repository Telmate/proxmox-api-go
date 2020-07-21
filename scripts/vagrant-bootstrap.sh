#!/usr/bin/env bash

export DEBIAN_FRONTEND=noninteractive

# ensure required utilities are installed
apt-get update
apt-get install -y software-properties-common git make gnupg2

# make sure the hostname can be resolved via /etc/hosts
PVE_IP=$(hostname -I | awk '{print $1}')
sed -i "s/127.0.1.1/$PVE_IP/" /etc/hosts

# add proxmox repository and its key
apt-add-repository 'deb http://download.proxmox.com/debian/pve buster pve-no-subscription'
wget -qO- http://download.proxmox.com/debian/proxmox-ve-release-6.x.gpg | apt-key add -

# update repositories and system
apt-get update
apt-get full-upgrade

# install proxmox packages
apt-get install -y proxmox-ve postfix open-iscsi

# don't scan for other operating systems
apt-get remove -y os-prober

# set root password so that we can use it to login to Proxmox API
sudo -i passwd <<EOF
root
root
EOF
