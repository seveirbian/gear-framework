import sys
# package need to be installed, pip install docker
import docker 
import time
import yaml
import os
import random
import subprocess
import signal
import shutil
import urllib2

test_pull = False

pwd = os.getcwd()

client = docker.from_env()

def empty_cache():
    # docker system prune -a
    p = subprocess.Popen("docker system prune -a", shell=True, 
                            stdout = subprocess.PIPE,
                            stdin = subprocess.PIPE)
    stdout, stderr = p.communicate("y\n")
    ret_code = p.wait()
    if ret_code != 0:
        print "fail to empty cache"
    # empty cache
    shutil.rmtree('/var/lib/gear/public/')
    os.mkdir('/var/lib/gear/public/')
    shutil.rmtree('/var/lib/gear/private/')
    os.mkdir('/var/lib/gear/private/')
    print "empty cache! \n"

def run_command(file):
    p = subprocess.Popen("python "+file+" yes", shell=True, 
                            stdout = subprocess.PIPE,
                            stdin = subprocess.PIPE)
    ret_code = p.wait()
    return ret_code

def check_image_num(image):
    generator = Generator(os.path.split(os.path.realpath(__file__))[0]+"/"+image+"/image_versions.yaml")

    docker_images = generator.generateFromProfile()

    images = docker_images[1][image]

    return len(images)

def check_gear_ready(image):
    req = urllib2.urlopen("http://202.114.10.146:9999/v2/"+image+"-gear/tags/list")
    image_info = req.read().split("\"tags\":[")
    image_info = image_info[1].split("]}\n")
    image_info = image_info[0]
    image_info = image_info.split(",")
    image_num = len(image_info)
    if image_num != check_image_num(image):
        return False
    return True

def check_gearmd_ready(image):
    req = urllib2.urlopen("http://202.114.10.146:9999/v2/"+image+"-gearmd/tags/list")
    image_info = req.read().split("\"tags\":[")
    image_info = image_info[1].split("]}\n")
    image_info = image_info[0]
    image_info = image_info.split(",")
    image_num = len(image_info)
    if image_num != check_image_num(image):
        return False
    return True

def check_docker_images_size():
    docker_images = os.path.join("/var/lib/docker/geargraphdriver")
    local_data = subprocess.check_output(['du','-ms', docker_images]).split()[0].decode('utf-8')
    print "Docker images size: ", local_data

def test_one_image(image):
    if test_pull == True:
        empty_cache()

        print "pull docker images from docker hub"
        step1_file = os.path.join(pwd, image, "pull_docker_images_from_dockerhub.py")
        if run_command(step1_file) != 0:
            print "fail step 1"
        print "push docker images to private registry"
        step2_file = os.path.join(pwd, image, "push_docker_images_to_private_registry.py")
        if run_command(step2_file) != 0:
            print "fail step 2"
        print "push docker images to back registry"
        step3_file = os.path.join(pwd, image, "push_docker_images_to_back_registry.py")
        if run_command(step3_file) != 0:
            print "fail step 3"
            
    else:
        empty_cache()

        print "pull docker images from private registry"
        step4_file = os.path.join(pwd, image, "pull_docker_images_from_private_registry.py")
        if run_command(step4_file) != 0:
            print "fail step 4"
        # print docker images size
        check_docker_images_size()
        print "run docker images"
        step5_file = os.path.join(pwd, image, "run_docker_images.py")
        if run_command(step5_file) != 0:
            print "fail step 5"

        # make gear images
        empty_cache()

        print "check gear images"
        while check_gear_ready(image) != True:
            time.sleep(0.1)

        # print "make gearmd image..."
        # print "first pull gear images from private registry"
        # step_pull_gear_file = os.path.join(pwd, image, "first_pull_gear_images_from_private_registry.py")
        # if run_command(step_pull_gear_file) != 0:
        #     print "fail pull gear images"
        # print "first run gear images with cache"
        # step_make_gear_file = os.path.join(pwd, image, "first_run_gear_images.py")
        # if run_command(step_make_gear_file) != 0:
        #     print "fail run gear images"
        # print "gearmd image made!!!"

        empty_cache()

        # first run without cache
        print "first pull gear images from private registry"
        step6_file = os.path.join(pwd, image, "first_pull_gear_images_from_private_registry.py")
        if run_command(step6_file) != 0:
            print "fail step 6"
        print "first run gear images without cache"
        step7_file = os.path.join(pwd, image, "first_run_gear_images_without_cache.py")
        if run_command(step7_file) != 0:
            print "fail step 7"

        empty_cache()

        # first run with cache
        print "first pull gear images from private registry"
        step8_file = os.path.join(pwd, image, "first_pull_gear_images_from_private_registry.py")
        if run_command(step8_file) != 0:
            print "fail step 8"
        print "first run gear images with cache"
        step9_file = os.path.join(pwd, image, "first_run_gear_images.py")
        if run_command(step9_file) != 0:
            print "fail step 9"

        # empty_cache()

        # print "check gearmd images"

        # while check_gearmd_ready(image) != True:
        #     time.sleep(0.1)

        # # second run without cache
        # print "second pull gear images form private registry"
        # step10_file = os.path.join(pwd, image, "second_pull_gear_images_from_private_registry.py")
        # if run_command(step10_file) != 0:
        #     print "fail step 10"
        # print "second run gear images without cache"
        # step11_file = os.path.join(pwd, image, "second_run_gear_images_without_cache.py")
        # if run_command(step11_file) != 0:
        #     print "fail step 11"

        # empty_cache()

        # # second run with cache
        # print "second pull gear images form private registry"
        # step12_file = os.path.join(pwd, image, "second_pull_gear_images_from_private_registry.py")
        # if run_command(step12_file) != 0:
        #     print "fail step 12"
        # print "second run gear images with cache"
        # step13_file = os.path.join(pwd, image, "second_run_gear_images.py")
        # if run_command(step13_file) != 0:
        #     print "fail step 13"
    

class Generator:
    
    def __init__(self, profilePath=""):
        self.profilePath = profilePath
    
    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"
        
        with open(self.profilePath, 'r') as f:
            self.images = yaml.load(f, Loader=yaml.FullLoader)

        return self.images

if __name__ == "__main__":

    if len(sys.argv) == 2:
        test_pull = True

    generator = Generator(os.path.split(os.path.realpath(__file__))[0]+"/images.yaml")

    docker_images = generator.generateFromProfile()

    images = docker_images[0]["images"]

    for image in images:
        test_one_image(image)
