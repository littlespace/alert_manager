import React, { useContext } from "react";
import styled from "styled-components";

import ClearRoundedIcon from "@material-ui/icons/ClearRounded";
import HorizontalSplitRoundedIcon from "@material-ui/icons/HorizontalSplitRounded";

import { PRIMARY, SECONDARY } from "../../styles/styles";
import { ROW_SELECT_ACTIONS } from "../../library/utils";
import { TableContext } from "../contexts/TableContext";

const HeaderCell = styled.th`
  padding: 10px;
  font-size: medium;
  font-weight: 600;
  background-color: ${SECONDARY};
  border: 2px solid ${PRIMARY};
`;

const Row = styled.tr`
  text-align: center;
`;

function getGroupIcon(column) {
  return (
    <div {...column.getGroupByToggleProps()} style={{ cursor: "pointer" }}>
      {column.isGrouped ? <ClearRoundedIcon /> : <HorizontalSplitRoundedIcon />}
    </div>
  );
}

function handleOnClick(column, rowSelectDispatch, page) {
  if (column.id !== "selection") {
    return null;
  }

  if (column.allSelected) {
    // Currently all are selected, user wants to unselect ALL rows (data == all rows)
    column.allSelected = false;
    rowSelectDispatch({ type: ROW_SELECT_ACTIONS.UNSELECT_ALL, rows: page });
  } else if (!column.allSelected) {
    // currently unselect, so select all in the viewport (e.g current page)
    column.allSelected = true;
    rowSelectDispatch({ type: ROW_SELECT_ACTIONS.SELECT_ALL, rows: page });
  }
}

function HeaderCells(headers) {
  const { rowSelectDispatch, page } = useContext(TableContext);

  const headerCells = [];
  headers.map(column =>
    headerCells.push(
      <HeaderCell
        {...column.getHeaderProps()}
        onClick={() => handleOnClick(column, rowSelectDispatch, page)}
      >
        {column.render("Header")}
        {column.canGroupBy ? getGroupIcon(column) : null}
      </HeaderCell>
    )
  );

  return headerCells;
}

function HeaderRow() {
  const { headerGroups } = useContext(TableContext);

  return headerGroups.map(headerGroup => (
    <Row {...headerGroup.getHeaderGroupProps()}>
      {HeaderCells(headerGroup.headers)}
    </Row>
  ));
}

function TableHeader() {
  return (
    <thead>
      <HeaderRow />
    </thead>
  );
}

export default TableHeader;
