import sys
# package need to be installed, pip install docker
import docker 
import time
import yaml
import os

auto = False

private_registry = "202.114.10.146:9999/"
suffix = "-gear"

class Puller:

    def __init__(self, images):  
        self.images_to_pull = images

    def check(self):
        # detect whether the file exists, if true, delete it
        if os.path.exists("./images_pulled.txt"):
            os.remove("./images_pulled.txt")
    
    def pull(self):
        self.check()

        client = docker.from_env()
        # if don't give a tag, then all image under this registry will be pulled
        repos = self.images_to_pull[0]["repo"]

        for repo in repos:
            tags = self.images_to_pull[1][repo]
            for tag in tags:
                print "start pulling: ", repo, ":", tag

                # get present time
                startTime = time.time()

                # get present net data
                cnetdata = get_net_data()

                # pull images
                try:
                    image_pulled = client.images.pull(repository=private_registry+repo+suffix, tag=str(tag))

                    # print pull time
                    finishTime = time.time() - startTime

                    print "finished in " , finishTime, "s\n"

                    # get image's size
                    size = image_pulled.attrs[u'Size'] / 1000000.0
                    print "image size: ", size

                    print "pull data: ", get_net_data() - cnetdata

                    # record the image and its pulling time
                    self.record(private_registry+repo+suffix, tag, finishTime, size)

                except docker.errors.NotFound:
                    print private_registry+repo+suffix + " not found...\n\n"
                except docker.errors.ImageNotFound:
                    print private_registry+repo+suffix + " image not fount...\n\n"

                if auto != True: 
                    raw_input("Next?")

    def record(self, repo, tag, time, size):
        with open("./images_pulled.txt", "a") as f:
            f.write("repo: "+str(repo)+" tag: "+str(tag)+" time: "+str(time)+" size: "+str(size)+"\n")

class Generator:
    
    def __init__(self, profilePath=""):
        self.profilePath = profilePath
    
    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"
        
        with open(self.profilePath, 'r') as f:
            self.images = yaml.load(f)

        return self.images

def get_net_data():
    netCard = "/proc/net/dev"
    fd = open(netCard, "r")

    for line in fd.readlines():
        if line.find("enp0s3") >= 0:
            field = line.split()
            data = float(field[1]) / 1024.0 / 1024.0

    fd.Close()
    return data

if __name__ == "__main__":

    if len(sys.argv) == 2:
        auto = True

    generator = Generator(os.path.split(os.path.realpath(__file__))[0]+"/image_versions.yaml")

    images = generator.generateFromProfile()

    puller = Puller(images)

    puller.pull()