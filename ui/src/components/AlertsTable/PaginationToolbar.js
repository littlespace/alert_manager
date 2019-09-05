import React from 'react';
import styled from 'styled-components';

import FirstPageOutlinedIcon from '@material-ui/icons/FirstPageOutlined';
import LastPageOutlinedIcon from '@material-ui/icons/LastPageOutlined';
import NavigateBeforeOutlinedIcon from '@material-ui/icons/NavigateBeforeOutlined';
import NavigateNextOutlinedIcon from '@material-ui/icons/NavigateNextOutlined';
import PageviewOutlinedIcon from '@material-ui/icons/PageviewOutlined';

import { PRIMARY, SECONDARY, HIGHLIGHT } from '../../styles/styles';

const PAGE_OPTIONS = [10, 50, 100];

const Pagination = styled.div`
  background-color: ${SECONDARY};
  height: 45px;
  display: block;
  position: relative;
`;

const Input = styled.input`
  background-color: ${SECONDARY};
  color: ${HIGHLIGHT};
  border: 0px solid ${PRIMARY};
  border-bottom: 1px solid ${HIGHLIGHT};
  font-size: medium;
  width: 30px;
  ::placeholder {
    color: ${HIGHLIGHT};
  }
`;

const Select = styled.select`
  background-color: ${SECONDARY};
  color: ${HIGHLIGHT};
  font-size: 0.875em;
  font-weight: 400;
  ::ms-expand {
    color: ${SECONDARY};
  }
`

const PageIcon = styled.span`
    padding-right: 2px;
    display: inline-flex;
    vertical-align: middle;
`

const Buttons = styled.span`
  position: absolute;
  right: 300px;
  padding-top: 10px;
`

const Label = styled.span`
  position: absolute;
  right: 173px;
  padding-top: 13px; 
`

const NumberSelect = styled.span`
  position: absolute;
  right: 100px;
  padding-top: 10px;
`

const TableSize = styled.span`
  position: absolute;
  right: 40px;
  padding-top: 12px;
`
function PaginationToolbar({
  pageCount,
  gotoPage,
  nextPage,
  previousPage,
  canNextPage,
  canPreviousPage,
  pageOptions,
  setPageSize,
  pageSize,
  pageIndex,
}) {
  return (
    <Pagination>
      <Buttons>
        <FirstPageOutlinedIcon
          style={{ cursor: 'pointer' }}
          fontSize={'medium'}
          onClick={() => gotoPage(0)}
          disabled={!canPreviousPage}
        />
        <NavigateBeforeOutlinedIcon
          style={{ cursor: 'pointer' }}
          fontSize={'medium'}
          onClick={() => previousPage()}
          disabled={!canPreviousPage}
        />
        <NavigateNextOutlinedIcon
          style={{ cursor: 'pointer' }}
          fontSize={'medium'}
          onClick={() => nextPage()}
          disabled={!canNextPage}
        />
        <LastPageOutlinedIcon
          style={{ cursor: 'pointer' }}
          fontSize={'medium'}
          onClick={() => gotoPage(pageCount - 1)}
          disabled={!canNextPage}
        />
      </Buttons>
      <Label>
        Page {pageIndex + 1} of {pageOptions.length}
      </Label>
      <NumberSelect>
        <PageIcon>
          <PageviewOutlinedIcon />
        </PageIcon>
        <Input
          type="number"
          defaultValue={pageIndex + 1}
          onChange={e => {
            const page = e.target.value ? Number(e.target.value) - 1 : 0;
            gotoPage(page);
          }}
        />
      </NumberSelect>
      <TableSize>
        <Select
          value={pageSize}
          onChange={e => {
            setPageSize(Number(e.target.value));
          }}
        >
          {PAGE_OPTIONS.map(pageSize => (
            <option key={pageSize} value={pageSize}>
              {pageSize}
            </option>
          ))}
        </Select>
      </TableSize>
    </Pagination>
  );
}

export default PaginationToolbar;
