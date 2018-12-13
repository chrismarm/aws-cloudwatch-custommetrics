'''
Created on Dec 13, 2018

@author: Christian Marmolejo

Code to be executed in a EC2 instance, that collects metrics about memory and 
disk usage and send them to CloudWatch, indexed by the instance id where the app is running
'''

import boto3
import subprocess
import requests

cloudwatch = None
instanceId = None
MEGA = 1024
MB = 'Megabytes'
PERCENT = 'Percent'

def metrics():
    init()
        
    ''' Memory values '''
    memory = getMemoryValues()
    totalMemory = float(memory['MemTotal'])
    availableMemory = float(memory['MemAvailable'])
    usedMemory = totalMemory - availableMemory
    swapTotal = float(memory['SwapTotal'])
    swapFree = float(memory['SwapFree'])
    swapUsed = swapTotal - swapFree
    
    sendMetric('MemoryUsage', usedMemory/totalMemory * 100, PERCENT)
    sendMetric('MemoryUsed', usedMemory / MEGA, MB)
    sendMetric('MemoryAvailable', availableMemory / MEGA, MB)
    sendMetric('SwapUsage', swapUsed/swapTotal if swapTotal > 0 else 0, PERCENT)
    sendMetric('SwapUsed', swapUsed / MEGA, MB)
    
    ''' Disk values '''
    disk = getDiskValues()
    usedDisk = float(disk[0])
    freeDisk = float(disk[1])
    percentDisk = float(disk[2].strip('%'))
    
    sendMetric('DiskUsage', percentDisk, PERCENT)
    sendMetric('DiskUsed', usedDisk / MEGA, MB)
    sendMetric('DiskFree', freeDisk / MEGA, MB)

def init():
    global instanceId
    global cloudwatch
    regionId, instanceId = getInstanceMetadata()
    cloudwatch = boto3.client('cloudwatch', regionId)

def getInstanceMetadata():
    ''' Request for metadata '''
    r = requests.get("http://169.254.169.254/latest/dynamic/instance-identity/document")
    ''' Let's force an error if the http response is not OK '''
    r.raise_for_status()
    response_json = r.json()
    
    region = response_json.get('region')
    instanceId = response_json.get('instanceId')
    return region, instanceId

def getMemoryValues():
    result = dict()
    with open('/proc/meminfo', 'r') as f:
        for line in f:
            items = line.split()
            name = items[0].strip(':')
            value = items[1]
            result[name] = value
    return result

def getDiskValues():
    result = list()
    out = subprocess.Popen(['df', '-l'], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    stdout,_ = out.communicate()
    lines = stdout.split('\n')
    for line in lines:
        items = line.split()
        '''Only the root file system'''
        if items[5] == '/':
            result.append(items[2])
            result.append(items[3])
            result.append(items[4])
            return result
    
    
    

def sendMetric(metric, value, unit):
    metric = [
            {
                'MetricName': metric,
                    'Dimensions': [{
                        'Name': 'InstanceId',
                        'Value': instanceId
                    }],
                'Unit': unit,
                'Value': value
            },
        ]
    response = cloudwatch.put_metric_data(MetricData=metric, Namespace='EC2MemoryDisk')
    if not str(response['ResponseMetadata']['HTTPStatusCode']).startswith('2'):
        raise ValueError('Metric ' + metric + " could not be sent")

if __name__ == '__main__':
    metrics()