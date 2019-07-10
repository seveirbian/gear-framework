import sys
# package need to be installed, pip install docker
import docker
import time
# package need to be installed, pip install pyyaml
import yaml
import os

class Detector:

    def __init__(self, images):
        self.images_to_push = images
        self.totalImageVirtualSize = 0

    def check(self):
        # detect whether the file exists, if true, delete it
        if os.path.exists("./images_detected.txt"):
            os.remove("./images_detected.txt")

    def detect(self):
        self.check()
        client = docker.from_env()

        # if don't give a tag, then all image under this registry will be pulled
        repos = self.images_to_push[0]["repo"]

        for repo in repos:
            tags = self.images_to_push[1][repo]

            for tag in tags:
                # get present time
                startTime = time.time()

                # detect images
                try:
                    image = client.images.get(name=repo+":"+str(tag))
                    virtualSize = image.attrs[u'VirtualSize']
                    self.totalImageVirtualSize += virtualSize

                    # print pull time
                    finishTime = time.time() - startTime
                    print "finished in " , finishTime, "s\n"

                    # record the image and its size
                    self.record(repo, tag, virtualSize)
                except docker.errors.NotFound:
                    print repo + ":" + str(tag) + " not found...\n\n"
                    self.record(repo, tag, "notfound")

    def record(self, repo, tag, virtualSize):
        with open("./images_detected.txt", "a") as f:
            f.write("repo: "+str(repo)+" tag: "+str(tag)+" viatualSize: " + str(virtualSize/1024.0/1024.0) + " MB\n")

class Generator:

    def __init__(self, profilePath=""):
        self.profilePath = profilePath

    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"

        with open(self.profilePath, 'r') as f:
            self.images_to_detect = yaml.load(f)

        return self.images_to_detect


if __name__ == "__main__":
    if len(sys.argv) >= 3 or len(sys.argv) <= 1:
        exit()

    generator = Generator(sys.argv[1])

    images_to_detect = generator.generateFromProfile()

    detector = Detector(images_to_detect)

    detector.detect()

    detector.record("totalImage", "", detector.totalImageVirtualSize)

    print "totalVirtualSize: ", str(detector.totalImageVirtualSize/1024.0/1024.0), " MB\n"