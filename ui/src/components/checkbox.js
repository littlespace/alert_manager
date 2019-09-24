import React from "react";
import PropTypes from "prop-types";

import CheckBoxOutlineBlankOutlinedIcon from "@material-ui/icons/CheckBoxOutlineBlankOutlined";
import CheckBoxOutlinedIcon from "@material-ui/icons/CheckBoxOutlined";
import IndeterminateCheckBoxOutlinedIcon from "@material-ui/icons/IndeterminateCheckBoxOutlined";

export function MultiSelectCheckbox({ selected, partial, ...props }) {
  let checkbox = null;
  if (partial) {
    checkbox = <IndeterminateCheckBoxOutlinedIcon fontSize={props.fontSize} />;
  } else if (selected) {
    checkbox = <CheckBoxOutlinedIcon fontSize={props.fontSize} />;
  } else {
    checkbox = <CheckBoxOutlineBlankOutlinedIcon fontSize={props.fontSize} />;
  }
  return <div>{checkbox}</div>;
}

MultiSelectCheckbox.propTypes = {
  selected: PropTypes.bool.isRequired,
  partial: PropTypes.bool
};
