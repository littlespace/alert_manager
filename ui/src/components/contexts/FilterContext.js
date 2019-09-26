import React, { createContext, useState } from "react";

let FilterContext;
const { Provider } = (FilterContext = createContext());

function FilterProvider(props) {
  const [filters, setFilters] = useState({});
  return (
    <Provider
      value={{
        filters: filters,
        setFilters: setFilters
      }}
    >
      {props.children}
    </Provider>
  );
}

export { FilterProvider, FilterContext };
