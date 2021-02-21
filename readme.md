# SGSync

I'm too lazy to whitelist my ip every time my box reboot as my ip is dynamic and i'm working on a cloud hosted VM. 

## Installation  

Create a config file in ~/.sgsync/config.json or pass it directly as argument when you run the executable.

sgsync config template:

```json
{
    "sgs": [
        {
            "id": "sg-a4f4264a",
            "port": 22,
            "comment": "for testing purposes"
        }
    ],
    "extra": { 
        "endpoint": "endpoint",
        "region": "eu-west-2"
    }
}
```