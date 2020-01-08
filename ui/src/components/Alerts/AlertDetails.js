import React from "react";
import styled from "styled-components";

import { SEVERITY_COLORS } from "../../library/utils";

const FIELDS = [
  "id",
  "start_time",
  "device",
  "owner",
  "entity",
  "site",
  "team",
  "source",
  "tags",
  "description"
];

const Wrapper = styled.div`
  display: grid;
  grid-row-gap: 0.5em;
  padding: 1em 0;
  align-items: center;
  grid-template-columns: repeat(4, minmax(auto, 350px));
`;

const Detail = styled.div`
  margin: 1em;
`;

const Description = styled.div`
  grid-column: 2 / 5;
  margin: 0 1em;
`;

const Title = styled.span`
  text-transform: uppercase;
  font-weight: bold;
  color: ${({ severity }) =>
    SEVERITY_COLORS[severity.toLowerCase()]["background-color"]};
`;

const ValueInline = styled.span`
  font-style: ${props => (props.missing ? "italic" : null)};
  font-weight: ${props => (props.missing ? "bolder" : null)};
  cursor: ${({ link }) => (link ? "pointer" : null)};

  :hover {
    font-weight: ${({ link }) => (link ? "bold" : null)};
  }
`;

function formatOptions(field, value) {
  // Field formating
  if (field.toLowerCase() === "start_time") {
    value = new Date(value * 1000).toLocaleString();
  }
  field = field.replace("_", " ");

  // value formating
  if (Array.isArray(value)) {
    value = value.join(", ");
  }

  return [field, value];
}

function handleValueOnClick(field, value) {
  switch (field.toLowerCase()) {
    case "device":
      window.open(`https://netbox.simulprod.com/search/?q=${value}`);
      break;
    default:
      return;
  }
}

function isLink(field) {
  return field.toLowerCase() === "device";
}

function getFieldItem(alert, field, value, idx) {
  [field, value] = formatOptions(field, value);
  let title = <Title severity={alert.severity}>{field}: </Title>;

  let valueObj = (
    <ValueInline
      missing={!value}
      link={isLink(field)}
      onClick={() => handleValueOnClick(field, value)}
    >
      {value || "Not Specified"}
    </ValueInline>
  );

  switch (field.toLowerCase()) {
    case "description":
      return (
        <Description key={idx}>
          {title}
          {valueObj}
        </Description>
      );
    default:
      return (
        <Detail key={idx}>
          {title}
          {valueObj}
        </Detail>
      );
  }
}

function AlertDetails({ alert }) {
  return (
    <Wrapper>
      {FIELDS.map((field, idx) => {
        const value = alert[field];
        return getFieldItem(alert, field, value, idx);
      })}
    </Wrapper>
  );
}
export default AlertDetails;
