import { createClient } from "@clickhouse/client";
import express from "express";

const app = express();
const port = 4000;
const host = "0.0.0.0";

// ClickHouse connection details
const db_host = process.env.CLICKHOUSE_HOST || "localhost";
const db_port = process.env.CLICKHOUSE_PORT || 8123;
const db_pass = process.env.CLICKHOUSE_PASSWORD || "otel";
const db_user = process.env.CLICKHOUSE_USER || "otel";

// Resource mapping
const resourceMapping = {
  "system.cpu.time": "cpu",
  "system.memory.usage": "memory",
};

// Create ClickHouse client
const client = createClient({
  url: `http://${db_host}:${db_port}`,
  password: db_pass,
  username: db_user,
});

function convertToDateTime(timestamp) {
  const date = new Date(timestamp * 1000);
  const formattedDate = date.toISOString().slice(0, 19).replace('T', ' ');
  return formattedDate;
}

app.get("/metrics", async (req, res) => {
  try {
    const start = convertToDateTime(Number(req.query.start));
    const end = convertToDateTime(Number(req.query.end));

    const result = await client.query({
      query: `
        SELECT
            MetricName AS resource,
            toUnixTimestamp(toStartOfMinute(TimeUnix)) AS minute,
            AVG(Value) AS avg_value
        FROM default.metrics
        WHERE TimeUnix >= toDateTime64('${start}', 9)
          AND TimeUnix <= toDateTime64('${end}', 9)
        GROUP BY
            MetricName,
            minute
        ORDER BY
            MetricName,
            minute;
      `,
      format: "JSONEachRow",
    });

    const rows = await result.json();

    const result2 = await client.query({
      query: `
        SELECT
            MetricName AS resource,
            MAX(Value) AS max_value
        FROM default.metrics
        WHERE TimeUnix >= toDateTime64('${start}', 9)
          AND TimeUnix <= toDateTime64('${end}', 9)
        GROUP BY
            MetricName
        ORDER BY
            MetricName;
      `,
      format: "JSONEachRow",
    });

    const rows2 = await result2.json();

    const resources = {};

    rows2.forEach(({ resource, max_value }) => {
      const key = resourceMapping[resource] || resource;

      if (!resources[key]) {
        resources[key] = { dates: [], values: [], max_value: null };
      }

      resources[key].max_value = max_value;
    });

    rows.forEach(({ resource, minute, avg_value }) => {
      const key = resourceMapping[resource] || resource;

      if (!resources[key]) {
        resources[key] = { dates: [], values: [], max_value: null };
      }

      resources[key].dates.push(minute);
      resources[key].values.push(avg_value);
    });

    return res.status(200).json(resources);

  } catch (error) {
    console.error("Error executing query:", error);
    res.status(500).json({ error: "Internal server error" });
  }
});


// Start the server
app.listen(port, () => {
  console.log(`API listening at http://${host}:${port}`);
});
