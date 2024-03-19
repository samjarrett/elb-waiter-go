# elb-waiter-go

A golang version of [samjarrett/elb-waiter](https://github.com/samjarrett/elb-waiter).

Useful at boot of an autoscaling EC2 that's associated with ELBv2 target groups,
to ensure that it's showing as healthy before signalling cloudformation / other
automation to proceed

## Usage

```bash
elb-waiter -instance-id [INSTANCE ID]
```
