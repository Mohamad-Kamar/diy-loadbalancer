# diy-loadbalancer

## Usage

### Build and Run

```bash
docker-compose up --build
```

### Test Round Robin

```bash
bash tests/integration/test_round_robin.sh
```

You should see responses from all three backends in turn.
