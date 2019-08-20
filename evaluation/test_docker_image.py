import sys
# package need to be installed, pip install docker
import docker 
import time
import yaml
import os
import random
import subprocess
import signal

pwd = os.getcwd()

client = docker.from_env()

def empty_cache():
    # docker system prune -a
    try:
        client.containers.prune()
    except:
        print "fail to prune containers!"
    try:
        client.images.prune()
    except:
        print "fail to prune images!"
    # empty cache
    shutil.rmtree('/var/lib/gear/public/')
    os.mkdir('/var/lib/gear/public/')
    shutil.rmtree('/var/lib/gear/private/')
    os.mkdir('/var/lib/gear/private/')
    print "empty cache! \n"

def run_command(file):
    p = subprocess.Popen("python "+file+" yes", shell=True, stdout=subprocess.STDOUT)
    ret_code = p.wait()
    return ret_code

def test_one_image(image):
    # 清空缓存
    empty_cache()

    # 从dockerhub下载镜像
    step1_file = os.path.join(pwd, image, "pull_docker_images_from_dockerhub.py")
    if run_command(step1_file) != 0:
        print "fail step 1"
    # 向private仓库push镜像
    step2_file = os.path.join(pwd, image, "push_docker_images_to_private_registry.py")
    if run_command(step2_file) != 0:
        print "fail step 2"
    # 向back_registry仓库push镜像
    step3_file = os.path.join(pwd, image, "push_docker_images_to_back_registry.py")
    if run_command(step3_file) != 0:
        print "fail step 3"

    # 清空缓存
    empty_cache()

    # 测试docker镜像的下载
    step4_file = os.path.join(pwd, image, "pull_docker_images_from_private_registry.py")
    if run_command(step4_file) != 0:
        print "fail step 4"
    # 测试docker镜像的运行
    step5_file = os.path.join(pwd, image, "run_docker_images.py")
    if run_command(step5_file) != 0:
        print "fail step 5"

    # 测试gear镜像的下载
    step6_file = os.path.join(pwd, image, "first_pull_gear_images_from_private_registry.py")
    if run_command(step6_file) != 0:
        print "fail step 6"
    # 测试gear镜像的第一次无缓存运行
    step7_file = os.path.join(pwd, image, "first_run_gear_images_without_cache.py")
    if run_command(step7_file) != 0:
        print "fail step 7"

    # 清空缓存
    empty_cache()

    # 测试gear镜像的下载
    step8_file = os.path.join(pwd, image, "second_pull_gear_images_from_private_registry.py")
    if run_command(step8_file) != 0:
        print "fail step 8"
    # 测试gear镜像的第一次无缓存运行
    step9_file = os.path.join(pwd, image, "second_run_gear_images_without_cache.py")
    if run_command(step9_file) != 0:
        print "fail step 9"

    # 清空缓存
    empty_cache()

    # 测试gear镜像的下载
    step10_file = os.path.join(pwd, image, "second_pull_gear_images_from_private_registry.py")
    if run_command(step10_file) != 0:
        print "fail step 10"
    # 测试gear镜像的第一次无缓存运行
    step11_file = os.path.join(pwd, image, "second_run_gear_images.py")
    if run_command(step11_file) != 0:
        print "fail step 11"

class Generator:
    
    def __init__(self, profilePath=""):
        self.profilePath = profilePath
    
    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"
        
        with open(self.profilePath, 'r') as f:
            self.images = yaml.load(f)

        return self.images

if __name__ == "__main__":

    if len(sys.argv) == 2:
        auto = True

    generator = Generator(os.path.split(os.path.realpath(__file__))[0]+"/images.yaml")

    docker_images = generator.generateFromProfile()

    images = docker_images[0]["images"]

    for image in images:
        test_one_image(image)
