# notes

1. python sdk needs to install docker and yaml packages
    - pip install docker
    - pip install pyyaml

2. private registry err: http: server gave HTTP response to HTTPS client
    1. Create or modify /etc/docker/daemon.json
        { "insecure-registries":["myregistry.example.com:5000"] }
    2. Restart docker daemon
        sudo service docker restart

3. 计算文件级共享的时候，忽略了字符设备、块设备、软连接等等，只对普通文件进行了检测
4. 计算启动时间时，对于需要其他容器配合的容器，没有测量(语言容器需要改进)