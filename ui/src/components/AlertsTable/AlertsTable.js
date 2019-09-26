import React, { useEffect, useContext } from "react";
import styled from "styled-components";

import { FilterContext } from "../contexts/FilterContext";
import { PRIMARY } from "../../styles/styles";
import { ROW_SELECT_ACTIONS, TABLE_ACTIONS } from "../../library/utils";
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

export default function AlertsTable({
  tableMutationState,
  tableMutationDispatch
}) {
  const {
    columns,
    getTableProps,
    state,
    rowSelectDispatch,
    rowSelectState
  } = useContext(TableContext);
  const { filters } = useContext(FilterContext);

  useEffect(() => {
    // Set allSelected to false to trigger a checkbox update for Selection Header Cell
    columns.find(c => c.id === "selection").allSelected = false;
    rowSelectDispatch({
      type: ROW_SELECT_ACTIONS.UNSELECT_ALL
    });

    if (tableMutationState.clearSelection) {
      // Cleanup the filters and row selects on unmount
      return () => {
        tableMutationDispatch({
          type: TABLE_ACTIONS.SET_CLEAR_SELECTION
        });
        rowSelectDispatch({ type: ROW_SELECT_ACTIONS.UNSELECT_ALL });
      };
    }
  }, [
    filters,
    state[0].pageSize,
    state[0].pageIndex,
    tableMutationState.clearSelection
  ]);

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
