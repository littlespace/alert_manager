import React from 'react';
import styled from 'styled-components';

import PageviewRoundedIcon from '@material-ui/icons/PageviewRounded';

import { getAlertFilterOptions } from '../../library/utils';
import { PRIMARY } from '../../styles/styles';
import FilterInput from './FilterInput';
import FilterMultiSelect from './FilterMultiSelect';
import FilterSelect from './FilterSelect';

// This name must match the column assesor field.
const MULTI_FILTERS = [
  'severity',
  'status',
  'device',
  'site',
  'source',
  'entity',
];

const Toolbar = styled.div`
  background-color: ${PRIMARY};
  padding: 25px;
  position: relative;
`;

const Icon = styled.span`
  display: inline-flex;
  position: absolute;
  top: 17px;
  right: 402px;
  vertical-align: middle;
`;

const GridStyle = styled.div`
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  grid-gap: 10px;
  padding-top: 45px;
  padding-bottom: 30px;
`;

function FilterToolbar({ alerts, ...props }) {
  return (
    <Toolbar>
      <Icon>
        <PageviewRoundedIcon />
      </Icon>
      <FilterInput
        filterType={'name'}
        placeholder={'Search by alert name...'}
        {...props}
      />
      <FilterSelect {...props} />
      <GridStyle>
        {MULTI_FILTERS.map(filterType => {
          return (
            <FilterMultiSelect
              filterType={filterType}
              options={getAlertFilterOptions(alerts, filterType)}
              placeholder={filterType}
              {...props}
            />
          );
        })}
      </GridStyle>
    </Toolbar>
  );
}

export default FilterToolbar;
