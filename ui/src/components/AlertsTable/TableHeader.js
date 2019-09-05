import React from 'react';
import styled from 'styled-components';

import ClearRoundedIcon from '@material-ui/icons/ClearRounded';
import HorizontalSplitRoundedIcon from '@material-ui/icons/HorizontalSplitRounded';

import { PRIMARY, SECONDARY } from '../../styles/styles';

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
    <div {...column.getGroupByToggleProps()} style={{ cursor: 'pointer' }}>
      {column.isGrouped ? <ClearRoundedIcon /> : <HorizontalSplitRoundedIcon />}
    </div>
  );
}

function getHeaderCells(headers) {
  const headerCells = [];
  headers.map(column =>
    headerCells.push(
      <HeaderCell {...column.getHeaderProps()}>
        {column.render('Header')}
        {column.canGroupBy ? getGroupIcon(column) : null}
      </HeaderCell>,
    ),
  );

  return headerCells;
}

function HeaderRow({ headerGroups }) {
  return headerGroups.map(headerGroup => (
    <Row {...headerGroup.getHeaderGroupProps()}>
      {getHeaderCells(headerGroup.headers)}
    </Row>
  ));
}

function TableHeader(props) {
  return (
    <thead>
      <HeaderRow {...props} />
    </thead>
  );
}

export default TableHeader;
