import React from "react";

import { MultiSelectCheckbox } from "../checkbox";

const COLUMNS = [
  {
    id: "selection",
    Header: props => getHeaderCheckbox(props),
    Cell: props => getCellCheckbox(props),
    allSelected: false
  },
  { accessor: "id", Header: "Id", show: false },
  { accessor: "history", Header: "History", show: false },
  { accessor: "severity", Header: "Severity", filter: SelectMultiColumnFilter },
  { accessor: "status", Header: "Status", filter: SelectMultiColumnFilter },
  { accessor: "name", Header: "Name" },
  { accessor: "site", Header: "Site", filter: SelectMultiColumnFilter },
  { accessor: "device", Header: "Device", filter: SelectMultiColumnFilter },
  { accessor: "entity", Header: "Entity", filter: SelectMultiColumnFilter },
  { accessor: "source", Header: "Source", filter: SelectMultiColumnFilter },
  { accessor: "last_active", Header: "Last Active" },
  { accessor: "start_date", Header: "Start Date", show: true },
  { accessor: "start_time", Header: "Start Time", show: true }
];

function SelectMultiColumnFilter(rows, id, filterValue) {
  let filteredRows = rows.filter(row => {
    return filterValue.includes(row.values[id]);
  });
  return filteredRows;
}

function getCellCheckbox({ row }) {
  let partial = false;
  if (row.canExpand) {
    // Grab all subrows that are selected
    const selectedRows = row.subRows.filter(row => row.isSelected);
    // If we have subrows selected and it DOES NOT match the length of subrows, partial == true
    partial = selectedRows.length
      ? selectedRows.length !== row.subRows.length
      : false;
  }

  return <MultiSelectCheckbox partial={partial} selected={row.isSelected} />;
}

function getHeaderCheckbox({ page, column, state }) {
  // If we have rows selected, determine if it's all rows on the current page
  // if it's all rows, return false for partial. If there are no rows selected
  // return false.
  const selectedRows = state[0].selectedRows;
  const partial = selectedRows.length
    ? selectedRows.length !== page.length
    : false;

  // check if any rows are aggregated (grouped)
  const rowsGrouped = !!page.find(row => row.isAggregated);
  return rowsGrouped ? null : (
    <MultiSelectCheckbox partial={partial} selected={column.allSelected} />
  );
}

export default COLUMNS;
