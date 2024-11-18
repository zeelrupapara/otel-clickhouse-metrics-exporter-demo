import React, { useState, useEffect } from "react";
import axios from "axios";
import { DateRangePicker } from "./components/Filters";
import { LineChart } from "./components/Charts";
import "./App.css";
const convertToUnixTimestamp = (dateTimeString) => {
  return Math.floor(new Date(dateTimeString).getTime() / 1000);
};

const getDefaultTimeRange = () => {
  const now = new Date();
  const oneHourAgo = new Date(now.getTime() - 1000 * 60 * 60);
  return [oneHourAgo, now];
};

const App = () => {
  const [cpuChartData, setCPUChartData] = useState([]);
  const [memoryChartData, setMemoryChartData] = useState([]);
  const [maxCPUUsage, setMaxUsage] = useState(0);
  const [maxMemoryUsage, setAvgUsage] = useState(0);

  useEffect(() => {
    const [defaultStartTime, defaultEndTime] = getDefaultTimeRange();
    fetchMetricsData(defaultStartTime, defaultEndTime);
  }, []);

  const fetchMetricsData = async (startTime, endTime) => {
    try {
      const response = await axios.get(
        `metrics?start=${convertToUnixTimestamp(
          startTime
        )}&end=${convertToUnixTimestamp(endTime)}`,
        {
          headers: {
            "Content-Type": "application/json",
            Accept: "application/json",
            "Access-Control-Allow-Origin": "*",
          },
        }
      );

      const { cpu, memory } = response.data;

      setCPUChartData(cpu ? [cpu.dates, cpu.values] : []);
      setMemoryChartData(memory ? [memory.dates, memory.values] : []);

      setMaxUsage(cpu ? cpu.max_value : 0);
      setAvgUsage(memory ? (memory.max_value / 1024 / 1024).toFixed(2) : 0);
    } catch (error) {
      console.error("Error fetching metrics data:", error);
    }
  };

  return (
    <div className="App">
      <DateRangePicker
        className="date-picker-container"
        onDateChange={fetchMetricsData}
      />

      <div className="chart-container">
      <div className="card">
            <p>Max CPU Usage</p>
            <h2>{maxCPUUsage}s</h2>
          </div>
        <div className="chart">
          {cpuChartData.length > 0 ? (
            <LineChart chartData={cpuChartData} title="CPU" unit="s" />
          ) : (
            <div className="no-data">No CPU data available</div>
          )}
        </div>
        <div className="card">
          <p>Max Memory Usage</p>
          <h2>{maxMemoryUsage}Gb</h2>
        </div>
        <div className="chart">
          {memoryChartData.length > 0 ? (
            <LineChart
              chartData={memoryChartData}
              title="Memory"
              unit="Bytes"
            />
          ) : (
            <div className="no-data">No Memory data available</div>
          )}
        </div>
      </div>
    </div>
  );
};

export default App;
