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

def test_one_image(image):
    empty_cache()

    step1_file = os.path.join(pwd, image, "pull_docker_images_from_dockerhub.py")
    if run_command(step1_file) != 0:
        print "fail step 1"
    step2_file = os.path.join(pwd, image, "push_docker_images_to_private_registry.py")
    if run_command(step2_file) != 0:
        print "fail step 2"
    step3_file = os.path.join(pwd, image, "push_docker_images_to_back_registry.py")
    if run_command(step3_file) != 0:
        print "fail step 3"

    empty_cache()

    step4_file = os.path.join(pwd, image, "pull_docker_images_from_private_registry.py")
    if run_command(step4_file) != 0:
        print "fail step 4"
    step5_file = os.path.join(pwd, image, "run_docker_images.py")
    if run_command(step5_file) != 0:
        print "fail step 5"

    step6_file = os.path.join(pwd, image, "first_pull_gear_images_from_private_registry.py")
    if run_command(step6_file) != 0:
        print "fail step 6"
    step7_file = os.path.join(pwd, image, "first_run_gear_images_without_cache.py")
    if run_command(step7_file) != 0:
        print "fail step 7"

    empty_cache()

    step8_file = os.path.join(pwd, image, "second_pull_gear_images_from_private_registry.py")
    if run_command(step8_file) != 0:
        print "fail step 8"
    step9_file = os.path.join(pwd, image, "second_run_gear_images_without_cache.py")
    if run_command(step9_file) != 0:
        print "fail step 9"

    empty_cache()

    step10_file = os.path.join(pwd, image, "second_pull_gear_images_from_private_registry.py")
    if run_command(step10_file) != 0:
        print "fail step 10"
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
