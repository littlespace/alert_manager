import React from "react";
import IconButton from "@material-ui/core/IconButton";
import Tooltip from "@material-ui/core/Tooltip";
import Archive from "@material-ui/icons/Archive";
import Block from "@material-ui/icons/Block";
import { withStyles } from "@material-ui/core/styles";

const defaultToolbarSelectStyles = {
  iconButton: {
    marginRight: "24px",
    top: "50%",
    display: "inline-block",
    position: "relative",
    transform: "translateY(-50%)"
  },

  icon: {
    color: "#424242"
  }
};

class CustomToolbarSelect extends React.Component {
  handleClick = () => {
    console.log("click! current selected rows", this.props.selectedRows);
  };

  render() {
    const { classes } = this.props;

    return (
      <div className={"custom-toolbar-select"}>
        <Tooltip title={"Change Status to CLEARED (not activated yet)"}>
          <IconButton className={classes.iconButton} onClick={this.handleClick}>
            <Archive className={classes.icon} />
          </IconButton>
        </Tooltip>
        <Tooltip title={"Change Status to SUPPRESSED (not activated yet)"}>
          <IconButton className={classes.iconButton} onClick={this.handleClick}>
            <Block className={classes.icon} />
          </IconButton>
        </Tooltip>
      </div>
    );
  }
}

export default withStyles(defaultToolbarSelectStyles, {
  name: "CustomToolbarSelect"
})(CustomToolbarSelect);
