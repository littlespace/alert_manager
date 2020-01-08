import React, { useContext } from "react";
import styled from "styled-components";

import ArrowDropDownOutlinedIcon from "@material-ui/icons/ArrowDropDownOutlined";
import ArrowRightIcon from "@material-ui/icons/ArrowRight";

import { HIGHLIGHT, PRIMARY, SECONDARY } from "../../styles/styles";
import { ROW_SELECT_ACTIONS } from "../../library/utils";
import { SEVERITY_COLORS } from "../../library/utils";

import { TableContext } from "../contexts/TableContext";

import ToolTip from "../ToolTip";

const StyledCell = styled.td`
  background-color: ${props =>
    // On grouping, the cell will be null
    props.value === null
      ? null
      : // Otherwise, if column is the severity column, set to the proper color
      props.columnId === "severity"
      ? SEVERITY_COLORS[props.value.toLowerCase()]["background-color"]
      : null};
  color: ${props => (props.columnId === "severity" ? PRIMARY : null)};
  cursor: ${props => (props.columnId === "selection" ? null : "pointer")};
  padding: 1em;

  tr:hover & {
    background-color: ${props =>
      // On grouping, the cell will be null
      props.value === null
        ? null
        : // Otherwise, if column is the severity column, set to the proper color
        props.columnId === "severity"
        ? SEVERITY_COLORS[props.value.toLowerCase()]["background-color"]
        : HIGHLIGHT};
    color: ${PRIMARY};
  }
`;

const Row = styled.tr`
  background-color: ${props => (props.row.isSelected ? HIGHLIGHT : SECONDARY)};
  color: ${props => (props.row.isSelected ? PRIMARY : null)};

  td:first-of-type {
    border-top-left-radius: 15px;
    border-bottom-left-radius: 15px;
  }

  td:last-of-type {
    border-top-right-radius: 15px;
    border-bottom-right-radius: 15px;
  }
`;

const Expanded = styled.span`
  display: inline-flex;
  margin-right: 2px;
  vertical-align: middle;
`;

function handleSelectionCellClick(row, rowSelectDispatch, prepareRow) {
  if (row.isSelected && row.canExpand) {
    // Row is a grouped row and selected; unselect all in the group
    row.subRows.forEach(row => {
      prepareRow(row);
      if (row.isSelected) {
        rowSelectDispatch({ type: ROW_SELECT_ACTIONS.UNSELECT_ROW, row: row });
      }
    });
  } else if (row.isSelected && !row.canExpand) {
    // Row is a single row and selected; unselect row
    rowSelectDispatch({ type: ROW_SELECT_ACTIONS.UNSELECT_ROW, row: row });
  } else if (!row.isSelected && row.canExpand) {
    // Row is a grouped row and NOT selected; select all subrows

    /* If you select the "select all" before the rows are expanded the rows do not have all the props
      e.g "row.toggleRowSelected". So our SELECT_ALL action will fail. We need to prepare each row before
      sending it off. */
    const unselectedRows = row.subRows.filter(row => !row.isSelected);
    row.subRows.forEach(row => {
      prepareRow(row);
      if (!row.isSelected) {
        rowSelectDispatch({ type: ROW_SELECT_ACTIONS.SELECT_ROW, row: row });
      }
    });
  } else if (!row.isSelected && !row.canExpand) {
    // Row is a single row and NOT selected; select row
    rowSelectDispatch({ type: ROW_SELECT_ACTIONS.SELECT_ROW, row: row });
  }
}

function handleCellClick(columnId, row, rowSelectDispatch, prepareRow) {
  // Grouped row that is NOT selection, so we don't want any action taken on it otherwise it prevents expanding the row
  if (row.canExpand === true && columnId !== "selection") {
    return null;
  }

  // All cells will open the alert details except for the selection column
  if (columnId === "selection") {
    handleSelectionCellClick(row, rowSelectDispatch, prepareRow);
  } else {
    return window.open(`/alert/${row.values.id}`);
  }
}

function getToolTip({ cell, ...props }) {
  const msg = cell.row.values.history.slice(-1)[0].event;
  const background =
    SEVERITY_COLORS[cell.row.values.severity.toLowerCase()][
      "background-color"
    ] || null;
  return (
    <ToolTip
      msg={msg}
      value={cell.value}
      background={background}
      position={"top"}
      {...props}
    />
  );
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

function Cell({ cell, row }) {
  const { rowSelectDispatch, prepareRow } = useContext(TableContext);
  const columnId = cell.column.id;
  const value = cell.value;

  return (
    <StyledCell
      columnId={columnId}
      value={value}
      {...cell.getCellProps({
        onClick: e =>
          handleCellClick(columnId, row, rowSelectDispatch, prepareRow)
      })}
    >
      {cell.isGrouped ? getGroupedCell(cell, row) : getAggregatedCell(cell)}
    </StyledCell>
  );
}

function getRow(row, prepareRow) {
  // You must run "prepareRow" on each row before rendering it. Otherwise the
  // render will fail. See API docs for react-table.
  return (
    prepareRow(row) || (
      <Row {...row.getRowProps()} row={row}>
        {row.cells.map((cell, index) => (
          <Cell cell={cell} row={row} key={row.values.id + index} />
        ))}
      </Row>
    )
  );
}

export default function TableBody() {
  const { page, prepareRow } = useContext(TableContext);
  return <tbody>{page.map(row => getRow(row, prepareRow))}</tbody>;
}
