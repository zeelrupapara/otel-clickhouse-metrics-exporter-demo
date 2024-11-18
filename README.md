# Demo
## clickhouse custom metrics exporter (opentelementory)

> Note: mostly things are just demo purpose, so code is not that complex for understanding how opentelemetry works

### Information
- `clickhouse-metrics-exporter`: is a custom metrics exporter for clickhouse
- `otelcol-dev`: custom collector environment inclucded binary
- `server`: node for get matrics from clickhouse
- `web`: visualize matrics

### How to use
#### 1) Start clickhouse database
```bash
   docker compose up -d
```

#### 2) Start Collector 
For matrics reciver we hostmetrics to get metrics of cpu and memory

```bash
   ./otelcol-dev/otelcol-dev --config="config.yaml"
```

#### 3) Start Server
```bash
cd server
npm install
node main.js
```
you can acess server in http://localhost:4000

#### 4) Start Web
```bash
cd web
npm install
npm start
```
you can acess web in http://localhost:3000