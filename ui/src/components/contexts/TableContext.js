import React, { createContext, useReducer } from "react";
import {
  useExpanded,
  useFilters,
  useGroupBy,
  usePagination,
  useRowSelect,
  useTable,
  useTableState
} from "react-table";
import { ROW_SELECT_ACTIONS } from "../../library/utils";

let TableContext;
const { Provider } = (TableContext = createContext());

function selectReducer(state, action) {
  const rowId = action.row ? action.row.values.id : null;

  switch (action.type) {
    case ROW_SELECT_ACTIONS.SELECT_ROW:
      state.rows.push(action.row);
      state.ids.push(rowId);

      action.row.toggleRowSelected(true);
      return { rows: state.rows, ids: state.ids };

    case ROW_SELECT_ACTIONS.UNSELECT_ROW:
      const filteredRows = state.rows.filter(row => row !== action.row);
      const filteredIds = state.ids.filter(id => id !== rowId);

      action.row.toggleRowSelected(false);
      return { rows: filteredRows, ids: filteredIds };

    case ROW_SELECT_ACTIONS.UNSELECT_ALL:
      state.rows.forEach(row => {
        try {
          row.toggleRowSelected(false);
        } catch (e) {
          if (e instanceof TypeError) {
            console.log(
              `Row element with ID:${row.values.id} was off the page due to \
              filtering, or pagination changes, so we didn't need to have it's \
              check, or highlight cleared. Do not fear though, it was still \
              removed from the selected rows.`
            );
          } else {
            throw new Error(e);
          }
        }
      });
      return { rows: [], ids: [] };

    case ROW_SELECT_ACTIONS.SELECT_ALL:
      const ids = [];
      action.rows.forEach(row => {
        row.toggleRowSelected(true);
        ids.push(row.values.id);
      });

      return { rows: action.rows, ids: ids };

    default:
      throw new Error("Incorrect action was selected for selecting rows");
  }
}

function TableProvider({ columns, data, ...props }) {
  /** Hoists the tableState from our react-table. The first arg is the
   *  initalState, the second is our overrides. `filters` will change from the
   *  cause a re=-render which will cause our "filter" to take place */
  const tableState = useTableState({ pageSize: 50 });

  const [rowSelectState, rowSelectDispatch] = useReducer(selectReducer, {
    rows: [],
    ids: []
  });

  const tableProps = useTable(
    {
      columns: columns,
      data: data,
      state: tableState
    },
    useFilters,
    useGroupBy,
    useExpanded,
    useRowSelect,
    usePagination
  );

  const exportProps = {
    rowSelectState,
    rowSelectDispatch,
    ...tableProps
  };

  return (
    <Provider
      value={{
        ...exportProps
      }}
    >
      {props.children}
    </Provider>
  );
}

export { TableProvider, TableContext };
