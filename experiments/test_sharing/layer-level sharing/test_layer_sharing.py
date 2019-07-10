import sys
# package need to be installed, pip install docker
import docker
import time
import yaml
import os
import hashlib

class Tester:

    def __init__(self, images):
        self.images_to_test = images

        # all layers
        self.layers = []

    def check(self):
        # detect whether the file exists, if true, delete it
        if os.path.exists("./images_tested.txt"):
            os.remove("./images_tested.txt")

    def test(self):
        self.check()

        client = docker.from_env()

        # if don't give a tag, then all image under this registry will be pulled
        repos = self.images_to_test[0]["repo"]

        for repo in repos:
            tags = self.images_to_test[1][repo]

            for tag in tags:

                # test images
                try:
                    image = client.images.get(name=repo+":"+str(tag))

                    # this is a array of upperdir in unicode 
                    upper_layers_in_unicode = image.attrs[u'GraphDriver'][u'Data'][u'UpperDir']

                    # only if layers_in_unicode only has ascii characters, use like this
                    upper_layers = str(upper_layers_in_unicode)
 
                    if image.attrs[u'GraphDriver'][u'Data'].has_key(u'LowerDir'):
                        lower_layers_in_unicode = image.attrs[u'GraphDriver'][u'Data'][u'LowerDir']
                        lower_layers = str(lower_layers_in_unicode)
                        # this is the array of path
                        layers = upper_layers.split(":") + lower_layers.split(":")
                    else:
                        layers = upper_layers.split(":")

                    # print image_files
                    self.layers += layers

                except docker.errors.NotFound:
                    print repo + " not found...\n\n"

        with open("./images_tested.txt", "a") as f:
            for layer in self.layers:
                f.write(layer+"\n")
            f.write("layer_num: "+str(len(self.layers))+"\n")

class Generator:

    def __init__(self, profilePath=""):
        self.profilePath = profilePath

    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"

        with open(self.profilePath, 'r') as f:
            self.images_to_push = yaml.load(f)

        return self.images_to_push


if __name__ == "__main__":
    if len(sys.argv) >= 3 or len(sys.argv) <= 1:
        exit()

    generator = Generator(sys.argv[1])

    images_to_test = generator.generateFromProfile()

    tester = Tester(images_to_test)

    tester.test()