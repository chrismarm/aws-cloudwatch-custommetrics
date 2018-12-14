### Cloud Watch custom metrics

These Python and Go components retrieve some metrics about a running EC2 instance and use AWS SDKs for both languages to push them into Cloud Watch. Both retrieve the id and the region of the EC2 instance from metadata endpoint. A role with enough privileges for Cloud Watch PutMetricData method must be created and associated to the instance before running these components.

* This `Python` script sends memory ('/proc/meminfo') and disk utilization ('df -l') metric values to Cloud Watch using `boto3` (AWS SDK for Python). The provided `userData` file contains the code that should be added to User Data field in instance launch configuration in order to install `pip` and `boto3` and download the script from S3.
* This `Go` command sends CPU parameters ('mpstat' iowait and idle parameters) and active TCP80 connections ('netstat -ant') of the instance using AWS SDK for Golang. The provided `userData` contains the needed code for the instance to install Git, Apache and Go.

Future improvements:
* Cloud Formation code to launch the instance, with the right role and security groups to allow communication