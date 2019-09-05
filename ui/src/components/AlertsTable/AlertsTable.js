import React from 'react';
import {
  useTable,
  useGroupBy,
  useFilters,
  useExpanded,
  usePagination,
  useTableState,
} from 'react-table';
import styled from 'styled-components';


import { PRIMARY } from '../../styles/styles';

import PaginationToolbar from './PaginationToolbar'
import TableHeaders from './TableHeader'
import TableBody from './TableBody'

const Table = styled.table`
  background-color: ${PRIMARY};
  font-size: 0.875rem;
  font-weight: 400;
  border-collapse: collapse;
  width: 100%;
`;

function AlertsTable(props) {
  /** Hoists the tableState from our react-table. The first arg is the 
   *  initalState, the second is our overrides. `filters` will change from the 
   *  cause a re=-render which will cause our "filter" to take place */
  const tableState = useTableState(
    { pageSize: 50 },
    { filters: props.filters },
  );

  const {
    getTableProps,
    headerGroups,
    prepareRow,
    page,
    ...tableProps
  } = useTable(
    {
      columns: props.columns,
      data: props.data,
      state: tableState,
    },
    useFilters,
    useGroupBy,
    useExpanded,
    usePagination,
  );

  return (
    <>
      {/* <div>
        <pre>
          <code>{JSON.stringify(state[0].filters, null, 2)}</code>
        </pre>
      </div> */}
      <Table {...getTableProps()}>
        <TableHeaders headerGroups={headerGroups} />
        <TableBody page={page} prepareRow={prepareRow} />
      </Table>
      <PaginationToolbar {...tableProps} />
   </>
  );
}


export default AlertsTable;