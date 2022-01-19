# Protected Entity Manager configuration directory structure

A Protected Entity Manager may take a config dir that contains configuration
info for Protected Entity types.  The structure of the directory is:

_config_dir_ 
|--s3config.json  
|--pes  
|  |--_pe type name_.pe.json

The _s3config.json_ file defines the S3 endpoint that will server transports for the
Protected Entities.  It is formatted as:

    {
    "host":"127.0.0.1",
    "port":9000,
    "accessKey":"accesskey",
    "secret":"secretkey",
    "region":"astrolabe",
    "http":true
    }

The accessKey and secret are used to generate pre-signed URLs to the S3 transport.

The fields of the _pe type name_.pe.json files are specific to the PE types.