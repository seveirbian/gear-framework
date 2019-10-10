import os
import hashlib

path = "/var/lib/docker/overlay2/"
hash_set = {}
storage = 0

def summary():
    print "files under %s\n"%path
    print "file number: %d\n"%len(hash_set)
    print "storage size: %d\n"%storage

def add_size(name):
    global storage
    fsize = os.path.getsize(name)
    fsize = fsize/float(1024*1024)
    storage += fsize

def insert(hash_value):
    if hash_set.has_key(hashlib) == False:
        hash_set[hash_value] = True
        return True
    return False

def calculate(name):
    print name

    f = open(name)
 
    the_hash = hashlib.md5()
     
    while True:
        d = f.read(8096)
        if not d:
          break
        the_hash.update(d)

    f.close()

    return the_hash.hexdigest()

def traverse(path):
    if os.path.exists(path):
        for root, dirs, files in os.walk(path):
            for file in files:
                name = os.path.join(root, file)
                if os.path.isfile(name) and os.path.islink(name):
                    hash_value = calculate(name)
                    if insert(hash_value) == True:
                        add_size(name)

if __name__ == "__main__":
    traverse(path)
    summary()