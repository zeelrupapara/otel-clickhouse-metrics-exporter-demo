import React from "react";
import UplotReact from "uplot-react";
import "uplot/dist/uPlot.min.css";

export const LineChart = ({chartData, title, unit}) => {  
  const options = {
    title: `${title} Usage`,
    width: 600,
    height: 400,
    axes: [
      { label: "Time (mins)" }, 
      { label: "Usage (%)" },
    ],
    series: [
      {},
      {
        label: `Average ${title} Usage (${unit})`,
        stroke: "blue",
        width: 2, 
      },
    ],
  };


  return (
    <div style={{ marginLeft: "100px", marginTop: "200px" }} >
      <UplotReact options={options} data={chartData} />
    </div>
  );
};