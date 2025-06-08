# Stateless Load Balancer

This is a lightweight, stateless load balancer written in Go (Golang). It is designed to forward traffic to a pool of targets, which are EC2 instances running in AWS. The load balancer dynamically manages its target pool based on instance health and configuration.

## Features

- **Dynamic Target Discovery**: 
    - Filter EC2 instances by tag name and value.
    - Automatically retrieve all instances in an Auto Scaling Group (ASG).
    - Retry mechanism for AWS API calls with exponential backoff.
- **Health Monitoring**: 
    - Continuously checks the health of target instances.
    - Automatically removes unhealthy instances from the load balancer pool.
    - Circuit breaker pattern to prevent cascading failures.
    - Configurable health check timeouts and intervals.
- **Connection Management**:
    - Connection pooling for better performance.
    - Automatic connection validation and cleanup.
    - Configurable connection timeouts and pool size.
- **Metrics and Monitoring**:
    - Prometheus metrics for:
        - Active connections
        - Health check latency
        - Failed/successful connections
        - Node health status
- **Reliability**:
    - Graceful shutdown with configurable timeout.
    - Proper cleanup of resources.
    - Error handling and recovery mechanisms.
- **Environment-Based Configuration**: 
    - All configurations are managed through environment variables for simplicity and flexibility.

## Configuration

The load balancer is configured using the following environment variables:

| Environment Variable           | Description                                                                 | Default Value |
|--------------------------------|-----------------------------------------------------------------------------|---------------|
| `PORT`                         | The port on which the load balancer listens for incoming traffic.           | 8080          |
| `HOST`                         | The host on which the load balancer listens for incoming traffic.           | 0.0.0.0       |
| `NODE_FILTER`                  | The way the LB filters the instances. `asg` or `tag`                        | asg           |
| `AWS_REGION`                   | The AWS region where the EC2 instances are located.                         | eu-west-1     |
| `TAG_NAME`                     | The tag name to filter EC2 instances. Required if `tag` is used.            | Name          |
| `TAG_VALUE`                    | The tag value to filter EC2 instances. Required if `tag` is used.           | worker-node   |
| `ASG_NAME`                     | The name of the Auto Scaling Group to retrieve instances from.              | worker-nodes  |
| `NODE_CHECK_PERIOD`            | The interval (in seconds) to get the instance IPs from AWS.                 | 10            |
| `HEALTH_CHECK_PERIOD`          | The interval (in seconds) for health checks on target instances.            | 2             |
| `HEALTH_CHECK_PATH`            | The path where the LB does health check requests.                           | /ping         |
| `HEALTH_CHECK_TIMEOUT`         | The timeout (in seconds) for health check requests.                         | 2             |
| `TARGET_PORT`                  | The port on which the target instances are listening.                       | 30080         |
| `POOL_SIZE`                    | Maximum number of connections in the connection pool.                       | 100           |
| `POOL_TIMEOUT`                 | Connection timeout (in seconds) for the connection pool.                    | 2             |
| `CIRCUIT_BREAKER_THRESHOLD`    | Number of failures before opening the circuit breaker.                      | 5             |
| `CIRCUIT_BREAKER_TIMEOUT`      | Time (in minutes) before the circuit breaker resets.                        | 5             |
| `SHUTDOWN_TIMEOUT`             | Timeout (in seconds) for graceful shutdown.                                 | 30            |
| `LOG_CONNECTIONS`              | Whether to log connection closure events. Useful for debugging.             | false         |

## Health Checks

The load balancer performs periodic health checks on all target instances. Unhealthy instances are automatically removed from the pool to ensure traffic is only forwarded to healthy targets. The health check system includes:

- HTTP-based health checks with configurable timeouts
- Circuit breaker pattern to prevent cascading failures
- Automatic retry mechanism for failed health checks
- Metrics for health check latency and success rates

## Connection Pooling

The load balancer implements connection pooling to improve performance and reduce overhead:

- Configurable pool size (default: 100 connections)
- Automatic connection validation
- Configurable connection timeouts
- Proper cleanup of stale connections

## Metrics

The load balancer exposes Prometheus metrics for monitoring:

- `load_balancer_active_connections`: Number of currently active connections
- `load_balancer_health_check_latency_seconds`: Latency of health checks
- `load_balancer_failed_connections_total`: Total number of failed connection attempts
- `load_balancer_successful_connections_total`: Total number of successful connection attempts
- `load_balancer_node_health_status`: Health status of each node (1 for healthy, 0 for unhealthy)

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

To configure the load balancer with custom settings:

```bash
export AWS_REGION=us-east-1
export NODE_FILTER=tag
export TAG_NAME=Environment
export TAG_VALUE=Production
export PORT=8080
export HEALTH_CHECK_PERIOD=30
export HEALTH_CHECK_PATH=/health
export HEALTH_CHECK_TIMEOUT=5
export TARGET_PORT=30080
export POOL_SIZE=200
export POOL_TIMEOUT=3
export CIRCUIT_BREAKER_THRESHOLD=3
export CIRCUIT_BREAKER_TIMEOUT=10
export SHUTDOWN_TIMEOUT=60

./stateless-load-balancer
```

## Graceful Shutdown

The load balancer supports graceful shutdown with a configurable timeout:

- Stops accepting new connections
- Waits for existing connections to complete
- Properly closes all resources
- Configurable timeout (default: 30 seconds)

## License

This project is licensed under the [MIT License](LICENSE).

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

