import os
import hashlib

path = "/var/lib/docker/overlay2/"
hash_set = {}
storage = 0.0
file_nums = 0

def summary():
    print "files under %s\n"%path
    print "unique file number: %d\n"%len(hash_set)
    print "file number: %d\n"%file_nums
    print "storage size: %d\n"%storage

def add_size(name):
    global storage
    fsize = os.path.getsize(name)*1.0
    fsize = fsize/float(1024*1024)
    storage += fsize
    print "current storage: %f\n"%storage

def insert(hash_value):
    if not hash_set.has_key(hash_value):
        hash_set[hash_value] = True
        return True
    return False

def calculate(name):
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
    global file_nums
    if os.path.exists(path):
        for root, dirs, files in os.walk(path):
            for file in files:
                name = os.path.join(root, file)
                if os.path.isfile(name) and (not os.path.islink(name)):
                    hash_value = calculate(name)
                    if insert(hash_value) == True:
                        add_size(name)
                    file_nums += 1

if __name__ == "__main__":
    traverse(path)
    summary()