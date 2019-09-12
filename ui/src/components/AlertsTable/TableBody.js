import React from 'react';
import styled from 'styled-components';

import ArrowDropDownOutlinedIcon from '@material-ui/icons/ArrowDropDownOutlined';
import ArrowRightIcon from '@material-ui/icons/ArrowRight';

import {
  PRIMARY,
  SECONDARY,
  HIGHLIGHT,
  INFO,
  WARN,
  CRITICAL,
} from '../../styles/styles';

const SEVERITYATTRS = {
  info: { 'background-color': INFO, color: 'black' },
  warn: { 'background-color': WARN, color: 'black' },
  critical: { 'background-color': CRITICAL, color: 'black' },
};

const Cell = styled.td`
  background-color: ${props =>
    // value of the cell is null (empty), we return null an use default value
    props.value === null
      ? null
      : // Otherwise, if column is the severity column, set to the proper color
      props.columnId === 'severity'
      ? SEVERITYATTRS[props.value.toLowerCase()]['background-color'] || null
      : null};
  color: ${props =>
    // value of the cell is null (empty), we return null an use default value
    props.value === null
      ? null
      : // Otherwise, if column is the severity column, set to the proper color
      props.columnId === 'severity'
      ? SEVERITYATTRS[props.value.toLowerCase()]['color'] || null
      : null};
  cursor: ${props => (props.columnId === 'details' ? 'pointer' : null)};
  padding: 10px;
  border-top: 2px solid ${PRIMARY};
  border-bottom: 2px solid ${PRIMARY};
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

function handleCellClick(columnID, rowID) {
  return columnID === 'details' ? window.open(`/alert/${rowID}`) : null;
}

function getAggregatedCell(cell) {
  /** If the cell is aggregated, use the Aggregated renderer for cell
     For cells with repeated values, render null Otherwise, 
    just render the regular cell */
  return cell.isAggregated
    ? cell.render('Aggregated')
    : cell.isRepeatedValue
    ? null
    : cell.render('Cell');
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
      {cell.render('Cell')} ({row.subRows.length})
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
      {...cell.getCellProps({
        onClick: () => handleCellClick(columnID, rowID),
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
