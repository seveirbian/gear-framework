import sys
# package need to be installed, pip install docker
import docker
import time
import yaml
import os

class Pusher:

    def __init__(self, images):
        self.images_to_push = images
        self.private_registry = "202.114.10.146:10000"

    def check(self):
        # detect whether the file exists, if true, delete it
        if os.path.exists("./images_pushed.txt"):
            os.remove("./images_pushed.txt")

    def push(self):
        self.check()
        client = docker.from_env()

        # if don't give a tag, then all image under this registry will be pulled
        repos = self.images_to_push[0]["repo"]

        for repo in repos:
            tags = self.images_to_push[1][repo]

            for tag in tags:
                # detect whether this image exists

                try:
                    client.images.get(name=self.private_registry+"/"+repo+":"+str(tag))
                except docker.errors.ImageNotFound:
                    print self.private_registry+"/"+repo, "is not found!\n"

                    try:
                        image = client.images.get(name=repo+":"+str(tag))
                        tagged = image.tag(repository=self.private_registry+"/"+repo, tag=str(tag))
                        if tagged is True:
                            print "tag ", repo+":"+str(tag), "to ", self.private_registry+"/"+repo+":"+str(tag), "\n"
                        else:
                            print "tag failed!\n"
                        
                    except docker.errors.ImageNotFound:
                        print repo, "is not found!\n"

                print "start pushing: ", self.private_registry+"/"+repo+":"+str(tag)
                # get present time
                startTime = time.time()

                # push images
                try:
                    messages = client.images.push(repository=self.private_registry+"/"+repo, tag=str(tag))

                    # print messages

                    # print pull time
                    finishTime = time.time() - startTime
                    print "finished in " , finishTime, "s\n"

                except docker.errors.NotFound:
                    print repo + " not found...\n\n"

class Generator:

    def __init__(self, profilePath=""):
        self.profilePath = profilePath

    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"

        with open(self.profilePath, 'r') as f:
            self.images_to_push = yaml.load(f, Loader=yaml.FullLoader)

        return self.images_to_push


if __name__ == "__main__":

    generator = Generator(os.path.split(os.path.realpath(__file__))[0]+"/image_versions.yaml")

    images_to_push = generator.generateFromProfile()

    pusher = Pusher(images_to_push)

    pusher.push()
