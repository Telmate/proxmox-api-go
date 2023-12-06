Vagrant.configure("2") do |config|
    config.vm.box = "generic/debian11"

    config.vm.boot_timeout = 1800
    config.vm.synced_folder ".", "/vagrant", disabled: true

    # forward proxmox API port
    config.vm.network "forwarded_port",
            guest: 8006,
            host_ip: "127.0.0.1",
            host: 8006

    # install and configure proxmox
    config.vm.provision "Bootstrap System", 
            type: "shell",
            privileged: true,
			path: './scripts/vagrant-bootstrap.sh'
    
    config.vm.provision "Import LXC Template",
            type: "shell",
            privileged: true,
            path: './scripts/vagrant-get-container-template.sh',
            run: "always"
    
    config.vm.provision "Download Cloud-Init Template",
            type: "shell",
            privileged: true,
            path: './scripts/vagrant-get-cloudinit-template.sh',
            run: "always"
    
    config.vm.provider :virtualbox do |vb|
        vb.memory = 2048
        vb.cpus = 2
        vb.customize ["modifyvm", :id, "--nested-hw-virt", "on"]
    end

    config.vm.provider :hyperv do |hv|
        hv.memory = 2048
        hv.cpus = 2
        hv.enable_virtualization_extensions = true
    end

    config.vm.provider :vmware do |vm|
        vm.vmx["memsize"] = "2048"
        vm.vmx["numvcpus"] = "2"
        vm.vmx["vhv.enable"] = "TRUE"
    end

    config.vm.provider :libvirt do |v, override|
        v.disk_bus = "virtio"
        v.driver = "kvm"
        v.video_vram = 8
        v.memory = 2048
        v.cpus = 2
    end
end
