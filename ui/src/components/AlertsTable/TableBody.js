import React from "react";
import styled from "styled-components";

import ArrowDropDownOutlinedIcon from "@material-ui/icons/ArrowDropDownOutlined";
import ArrowRightIcon from "@material-ui/icons/ArrowRight";

import {
  PRIMARY,
  SECONDARY,
  HIGHLIGHT,
  INFO,
  WARN,
  CRITICAL,
  ROBLOX
} from "../../styles/styles";
import ToolTip from "../ToolTip";
import HistoryItem from "../Alerts/HistoryItem";

const SEVERITYATTRS = {
  info: { "background-color": INFO, color: PRIMARY },
  warn: { "background-color": WARN, color: PRIMARY },
  critical: { "background-color": CRITICAL, color: PRIMARY }
};

const Cell = styled.td`
  background-color: ${props =>
    // value of the cell is null (empty), we return null an use default value
    props.value === null
      ? null
      : // Otherwise, if column is the severity column, set to the proper color
      props.columnId === "severity"
      ? SEVERITYATTRS[props.value.toLowerCase()]["background-color"]
      : null};
  color: ${props =>
    // value of the cell is null (empty), we return null an use default value
    props.value === null
      ? null
      : // Otherwise, if column is the severity column, set to the proper color
      props.columnId === "severity"
      ? SEVERITYATTRS[props.value.toLowerCase()]["color"]
      : null};
  cursor: ${props => (props.columnId === "details" ? "pointer" : null)};
  padding: 10px;
  border-top: 2px solid ${PRIMARY};
  border-bottom: 2px solid ${PRIMARY};
  :hover {
    color: ${props =>
      props.columnId === "details"
        ? SEVERITYATTRS[props.severity.toLowerCase()]["background-color"]
        : null};
  }
`;

const Row = styled.tr`
  background-color: ${SECONDARY};

  &:hover {
    background-color: ${HIGHLIGHT};
    color: ${PRIMARY};
  }

  td:first-child {
    border-top-left-radius: 15px;
    border-bottom-left-radius: 15px;
  }

  td:last-child {
    border-top-right-radius: 15px;
    border-bottom-right-radius: 15px;
  }
`;

const Expanded = styled.span`
  display: inline-flex;
  margin-right: 2px;
  vertical-align: middle;
`;

function getToolTip({ cell, ...props }) {
  const msg = cell.row.values.history.slice(-1)[0].event;
  const background =
    SEVERITYATTRS[cell.row.values.severity.toLowerCase()]["background-color"] ||
    null;
  return (
    <ToolTip msg={msg} value={cell.value} background={background} {...props} />
  );
}

function handleCellClick(columnID, rowID) {
  return columnID === "details" ? window.open(`/alert/${rowID}`) : null;
}

function getCellRenderer(cell) {
  let renderer = null;
  if (cell.column.id === "last_active") {
    // the render will pass the column and table props to the component
    renderer = getToolTip;
  } else {
    renderer = "Cell";
  }
  return renderer;
}

function getAggregatedCell(cell) {
  /** If the cell is aggregated, use the Aggregated renderer for cell
     For cells with repeated values, render null Otherwise, 
    just render the regular cell */
  const renderer = getCellRenderer(cell);
  return cell.isAggregated
    ? cell.render("Aggregated")
    : cell.isRepeatedValue
    ? null
    : cell.render(renderer);
}

function getGroupedCell(cell, row) {
  /* If it's a grouped cell, add an expander and row count
    The expanded cells break in 7.0.0.alpha.32 for react-table.
    The API changed so the "render('Cell') call now returns String(value)"
    Meaning you will get String(null) for the empty epanded cells which puts
    null all over the table. See here for
    more details: https://github.com/tannerlinsley/react-table/issues/1490 */
  return (
    <>
      <Expanded {...row.getExpandedToggleProps()}>
        {row.isExpanded ? <ArrowDropDownOutlinedIcon /> : <ArrowRightIcon />}
      </Expanded>
      {cell.render("Cell")} ({row.subRows.length})
    </>
  );
}

function getCell(cell, row) {
  const columnID = cell.column.id;
  const value = cell.value;
  const rowID = row.values.id;

  return (
    <Cell
      columnId={columnID}
      value={value}
      severity={cell.row.values.severity}
      {...cell.getCellProps({
        onClick: () => handleCellClick(columnID, rowID)
      })}
    >
      {cell.isGrouped ? getGroupedCell(cell, row) : getAggregatedCell(cell)}
    </Cell>
  );
}

function getRow(row, prepareRow) {
  // You must run "prepareRow" on each row before rendering it. Otherwise the
  // render will fail. See API docs for react-table.
  return (
    prepareRow(row) || (
      <Row {...row.getRowProps()}>
        {row.cells.map(cell => getCell(cell, row))}
      </Row>
    )
  );
}

function TableBody({ page, prepareRow }) {
  return <tbody>{page.map(row => getRow(row, prepareRow))}</tbody>;
}

export default TableBody;
