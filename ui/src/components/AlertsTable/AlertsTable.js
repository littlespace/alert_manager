import React, { useEffect, useContext } from "react";
import styled from "styled-components";

import { PRIMARY } from "../../styles/styles";
import { ROW_SELECT_ACTIONS } from "../../library/utils";
import { TableContext } from "../contexts/TableContext";
import PaginationToolbar from "./PaginationToolbar";
import SelectToolbar from "./SelectToolbar";
import TableBody from "./TableBody";
import TableHeaders from "./TableHeader";

const Table = styled.table`
  background-color: ${PRIMARY};
  font-size: 0.875rem;
  font-weight: 400;
  text-align: center;
  border-collapse: collapse;
  width: 100%;
`;

export default function AlertsTable() {
  const {
    columns,
    getTableProps,
    state,
    rowSelectDispatch,
    rowSelectState
  } = useContext(TableContext);

  useEffect(() => {
    // Set allSelected to false to trigger a checkbox update for Selection Header Cell
    columns.find(c => c.id === "selection").allSelected = false;
    rowSelectDispatch({
      type: ROW_SELECT_ACTIONS.UNSELECT_ALL
    });
  }, [state[0].pageSize, state[0].pageIndex]);

  return (
    <>
      {/* <div style={{ color: "black" }}>
        <pre>
          <code>{console.log("Rendering Table")}</code>
        </pre>
        <pre>
          <code>{JSON.stringify(filters, null, 2)}</code>
        </pre>
        <pre>
          <code>{JSON.stringify(rowSelectState.rows.length, null, 2)}</code>
        </pre>
        <pre>
          <code>{JSON.stringify(rowSelectState.ids, null, 2)}</code>
        </pre>
      </div> */}
      {rowSelectState.rows.length > 0 ? <SelectToolbar /> : null}
      <Table {...getTableProps()}>
        <TableHeaders />
        <TableBody />
      </Table>
      <PaginationToolbar />
    </>
  );
}
