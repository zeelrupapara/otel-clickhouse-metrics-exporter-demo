import React, { useState } from "react";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";

export const DateRangePicker = ({ onDateChange }) => {
  const [startDate, setStartDate] = useState(new Date());
  const [endDate, setEndDate] = useState(new Date());

  const selectDateTime = (startDate, endDate) => {
    setStartDate(startDate);
    setEndDate(endDate);
    onDateChange(startDate, endDate);
  };

  return (
    <div className="date-range-picker">
      <div className="date-picker">
        <label>Start Date:</label>
        <DatePicker
          selected={startDate}
          onChange={(date) => selectDateTime(date, endDate)}
          timeInputLabel="Time:"
          dateFormat="MM/dd/yyyy h:mm aa"
          showTimeInput
          className="date-input"
        />
      </div>
      <div className="date-picker">
        <label>End Date:</label>
        <DatePicker
          selected={endDate}
          onChange={(date) => selectDateTime(startDate, date)}
          timeInputLabel="Time:"
          dateFormat="MM/dd/yyyy h:mm aa"
          showTimeInput
          className="date-input"
        />
      </div>
    </div>
  );
};
