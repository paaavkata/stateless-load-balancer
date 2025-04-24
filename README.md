# Stateless Load Balancer

This is a lightweight, stateless load balancer written in Go (Golang). It is designed to forward traffic to a pool of targets, which are EC2 instances running in AWS. The load balancer dynamically manages its target pool based on instance health and configuration.

## Features

- **Dynamic Target Discovery**: 
    - Filter EC2 instances by tag name and value.
    - Automatically retrieve all instances in an Auto Scaling Group (ASG).
- **Health Monitoring**: 
    - Continuously checks the health of target instances.
    - Automatically removes unhealthy instances from the load balancer pool.
- **Environment-Based Configuration**: 
    - All configurations are managed through environment variables for simplicity and flexibility.

## Configuration

The load balancer is configured using the following environment variables:

| Environment Variable       | Description                                                                 |
|----------------------------|-----------------------------------------------------------------------------|
| `PORT`               | The port on which the load balancer listens for incoming traffic.                       |
| `HOST`               | The host on which the load balancer listens for incoming traffic.                        |
| `NODE_FILTER`               | The way the LB filters the instances. `asg` or `tag`                        |
| `AWS_REGION`               | The AWS region where the EC2 instances are located.                        |
| `TAG_NAME`          | (Optional) The tag name to filter EC2 instances. Required if `tag` is used for `NODE_FILTER`                          |
| `TAG_VALUE`         | (Optional) The tag value to filter EC2 instances. Required if `tag` is used for `NODE_FILTER`                          |
| `ASG_NAME`                 | (Optional) The name of the Auto Scaling Group to retrieve instances from. Required if `asg` is used for `NODE_FILTER` |
| `NODE_CHECK_PERIOD` | The interval (in seconds) to get the instance IPs from AWS. |
| `HEALTH_CHECK_PERIOD`    | The interval (in seconds) for health checks on target instances.           |
| `HEALTH_CHECK_PATH`     | The path where the LB does health check requests.                    |
| `TARGET_PORT`       | The port on which the load balancer listens for incoming traffic.          |
> **Note**: Either `TAG_NAME` and `TAG_VALUE` or `ASG_NAME` must be provided to define the target pool.

## Health Checks

The load balancer performs periodic health checks on all target instances. Unhealthy instances are automatically removed from the pool to ensure traffic is only forwarded to healthy targets.

## Usage

1. Clone the repository:
     ```bash
     git clone https://github.com/paaavkata/stateless-load-balancer
     cd stateless-load-balancer
     ```

2. Build the project:
     ```bash
     ./build_push.sh
     ```

3. Set the required environment variables.

4. Run the load balancer:
     ```bash
     ./stateless-load-balancer-arm64
     # or
     ./stateless-load-balancer-amd64
     ```

## Example

To configure the load balancer to forward traffic to instances with a specific tag:

```bash
export AWS_REGION=us-east-1
export TARGET_TAG_NAME=Environment
export TARGET_TAG_VALUE=Production
export LOAD_BALANCER_PORT=8080
export HEALTH_CHECK_INTERVAL=30
export HEALTH_CHECK_TIMEOUT=5

./stateless-load-balancer
```

## License

This project is licensed under the [MIT License](LICENSE).

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

