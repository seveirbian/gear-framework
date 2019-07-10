import sys
# package need to be installed, pip install docker
import docker 
import time
import yaml
import os

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

                # pull images
                try:
                    image_pulled = client.images.pull(repository=repo, tag=str(tag))

                    # print pull time
                    finishTime = time.time() - startTime

                    print "finished in " , finishTime, "s\n"

                    # record the image and its pulling time
                    self.record(repo, tag, finishTime)

                except docker.errors.NotFound:
                    print repo + " not found...\n\n"
                except docker.errors.ImageNotFound:
                    print repo + " image not fount...\n\n"

    def record(self, repo, tag, time):
        with open("./images_pulled.txt", "a") as f:
            f.write("repo: "+str(repo)+" tag: "+str(tag)+" time: "+str(time)+"\n")

class Generator:
    
    def __init__(self, profilePath=""):
        self.profilePath = profilePath
    
    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"
        
        with open(self.profilePath, 'r') as f:
            self.images_to_pull = yaml.load(f)

        return self.images_to_pull


if __name__ == "__main__":
    if len(sys.argv) >= 3 or len(sys.argv) <= 1:
        exit()

    generator = Generator(sys.argv[1])

    images_to_pull = generator.generateFromProfile()

    puller = Puller(images_to_pull)

    puller.pull()