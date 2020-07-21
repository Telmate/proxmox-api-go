Vagrant.configure("2") do |config|
    config.vm.box = "debian/buster64"

    config.vm.boot_timeout = 1800
    config.vm.synced_folder ".", "/vagrant", disabled: true

    # forward proxmox API port
    config.vm.network "forwarded_port",
            guest: 8006,
            host_ip: "127.0.0.1",
            host: 8006

    # install and configure proxmox
    config.vm.provision "shell",
            privileged: true,
			path: './scripts/vagrant-bootstrap.sh'
    
    config.vm.provider :virtualbox do |vb|
        vb.memory = 2048
        vb.cpus = 2
    end

    config.vm.provider :libvirt do |v, override|
        v.disk_bus = "virtio"
        v.driver = "kvm"
        v.video_vram = 8
        v.memory = 2048
        v.cpus = 2
    end
end
