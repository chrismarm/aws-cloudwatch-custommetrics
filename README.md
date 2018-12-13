### Cloud Watch custom metrics

This Python script is thought to be run in a EC2 instance and send metric values about memory and disk usage to Cloud Watch using `boto3` (AWS SDK for Python)

The provided `userData` file contains the code that should be added to User Data field in instance launch configuration in order to install `pip` and `boto3` and download the script from S3.

Future improvements:
* Golang script to provide more metrics
* Cloud Formation code to launch the instance