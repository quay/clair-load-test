# Clair Load Testing

Tool to run HTTP benchmarking tests on Clair v4.

## **Prerequisites**
### **Running on openshift cluster**
* A running instance of Clair (to test).
### **Running on local machine (Optional)**
* A running instance of Clair (to test).
* `clairctl` installed on your system (https://github.com/quay/clair/releases).
## **Installation**

```
make
```
> **NOTE**: You might want to add **sudo** to this step as it involves creating `clair-load-test` binary in your $PATH as well as uploading it to the remote image registry.

## **Usage**
### **Load Phase**
Run the `assets/image_load.sh` script with the required parameters to create tags in the repo which can be used for the testing later.

### **Example**
```
START=1 END=10000 LAYERS=5 IMAGES="quay.io/clair-load-test/mysql:8.0.25" LOAD_REPO="quay.io/vchalla/clair-load-test" RATE=20 bash image_load.sh
```
This will create 1 to 10000 tags of images specified in the load repo which will be helpful for our testing later.
> **NOTE**: One-time Activity. Use it if you want to specify either of `CLAIR_TEST_REPO_PREFIX/testrepoprefix` options in the Run phase.

### **Run Phase**
### **Usage on openshift platform**
Deploy `assets/clair-config.yaml` on to your openshift cluster.
### **Envs**
* `CLAIR_TEST_HOST` - String indicating clair host to perform testing.
* `CLAIR_TEST_CONTAINERS` - String with comma separated list of conatiner images.
* `CLAIR_TEST_RUNID`(Optional) - String specifying the desired RUNID of the test run.
* `CLAIR_TEST_PSK` - Psk string which can be found at `~/clair/config.yaml` in the clair app pod.
* `CLAIR_TEST_REPO_PREFIX` - String indicating the test repo prefix.
* `CLAIR_TEST_ES_HOST` - String indicating the ES instance host.
* `CLAIR_TEST_ES_PORT` - Indicates the port number of ES instance.
* `CLAIR_TEST_ES_INDEX` - String indicating the ES index to upload the results.
* `CLAIR_TEST_INDEX_REPORT_DELETE` - Boolean flag to indicate the index reports deletion at the end of the test run.
* `CLAIR_TEST_HIT_SIZE` - Indicates the total amount of requests to hit the system with.
* `CLAIR_TEST_LAYERS` - One among [5, 10, 15, 20, 25, 30, 35, 40] to pull image manifests with those many layers for testing. Valid only when pulling manifests from remote repository (i.e. using **CLAIR_TEST_REPO_PREFIX**) instead of using **CLAIR_TEST_CONTAINERS** option.
* `CLAIR_TEST_CONCURRENCY` - Indicates the rate(concurrency) at which the requests hits must happen in parallel.

Once triggered it will create a job in the specified namespace and will start running the tests with above mentioned values.

### **Example Usage**
Create a yaml similar to `assets/clair-config.yaml` and apply it to a desired namespace in your openshift cluster.
```
oc apply -f ~/assets/clair-config.yaml
```
Once deployed you should be able to see the details of the run in the pod logs with results getting logged and finally indexed to the target elastic search index.

### **Usage on Local Machine**
```
NAME:
   clair-load-test - A command-line tool for stress testing clair v4.

USAGE:
   clair-load-test [global options] command [command options] [arguments...]

VERSION:
   0.0.1

DESCRIPTION:
   A command-line tool for stress testing clair v4.

COMMANDS:
   report       clair-load-test report
   createtoken  createtoken --key sdfvevefr==
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -D             print debugging logs (default: false)
   -W             quieter log output (default: false)
   --help, -h     show help
   --version, -v  print the version
```

### Report
```
NAME:
   clair-load-test report - clair-load-test report

USAGE:
   clair-load-test report [command options] [arguments...]

DESCRIPTION:
   request reports for named containers

OPTIONS:
   --host value            --host localhost:6060/ (default: "http://localhost:6060/") [$CLAIR_TEST_HOST]
   --runid value           --runid f519d9b2-aa62-44ab-9ce8-4156b712f6d2 (default: "14484a83-abba-483c-9b66-3b5ce93b4088") [$CLAIR_TEST_RUNID]
   --containers value      --containers ubuntu:latest,mysql:latest [$CLAIR_TEST_CONTAINERS]
   --testrepoprefix value  --testrepoprefix quay.io/vchalla/clair-load-test:mysql_8.0.25 [$CLAIR_TEST_REPO_PREFIX]
   --psk value             --psk secretkey [$CLAIR_TEST_PSK]
   --eshost value          --eshost eshosturl [$CLAIR_TEST_ES_HOST]
   --esport value          --esport esport [$CLAIR_TEST_ES_PORT]
   --esindex value         --esindex esindex [$CLAIR_TEST_ES_INDEX]
   --delete                --delete (default: false) [$CLAIR_TEST_INDEX_REPORT_DELETE]
   --hitsize value         --hitsize 100 (default: 25) [$CLAIR_TEST_HIT_SIZE]
   --layers value          --layers 10 (default: 5) [$CLAIR_TEST_LAYERS]
   --concurrency value     --concurrency 50 (default: 10) [$CLAIR_TEST_CONCURRENCY]
   --help, -h              show help
```

### **Example Usage**
Processes the below list of containers and executes tests at rate of 10rps with 25 HTTP requests in total.
```
clair-load-test -D report --containers="quay.io/clair-load-test/ubuntu:xenial,quay.io/clair-load-test/ubuntu:focal,quay.io/clair-load-test/ubuntu:impish,quay.io/clair-load-test/ubuntu:trusty" --hitsize=25 --concurrency=10 --delete=true --host=http://example-registry-clair-app-quay-enterprise.apps.vchalla-clair-test.perfscale.devcluster.openshift.com --psk=RUZMTEVxMFI2QmVTRnhhNG5VUTF0ZVJZb1hLeTYwY20= --eshost="https://search-perfscale-dev-chmf5l4sh66lvxbnadi4bznl3a.us-west-2.es.amazonaws.com" --esport="443" --esindex="clair-test-index"
```

Gets the list of manifests from the test repo(created during load phase) which is specified through the `--testrepoprefix` option and runs the test at a rate of 10rps with 25 requests in total.
```
clair-load-test -D report --hitsize=25 --layers=5 --concurrency=10 --delete=true --host=http://example-registry-clair-app-quay-enterprise.apps.vchalla-clair-test.perfscale.devcluster.openshift.com --psk=RUZMTEVxMFI2QmVTRnhhNG5VUTF0ZVJZb1hLeTYwY20= --testrepoprefix="quay.io/vchalla/clair-load-test:mysql_8.0.25" --eshost="https://search-perfscale-dev-chmf5l4sh66lvxbnadi4bznl3a.us-west-2.es.amazonaws.com" --esport="443" --esindex="clair-test-index"
```
> **NOTE**: Both `--containers` and `--testrepoprefix` options are mutually exclusive.