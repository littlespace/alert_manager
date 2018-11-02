import React from "react";
import IconButton from "@material-ui/core/IconButton";
import Tooltip from "@material-ui/core/Tooltip";
import Archive from "@material-ui/icons/Archive";
import Block from "@material-ui/icons/Block";
import { withStyles } from "@material-ui/core/styles";

const styles = {
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

  constructor(props){
    super(props);
    this.classes = this.props.classes;
    this.api = this.props.api;
    this.idx = this.props.idx;
  }

//   selectedRowsToList(rows) {

  
//     let rows_ids = []

//     for( let i in rows ) {
//         for( let y in columns ) {
//             alert.push(data[i][columns[y].label])   
//         }
//         alerts.push(alert)
//     }
//     return alerts
// } 
//   }

  updateAlertsStatusSuppressed = (status) => {

    for(var i in this.props.selectedRows.data) {
        let row_id = this.props.selectedRows.data[i].dataIndex
        // console.log(`Alert Row ${row_id} / ${this.idx[row_id]}`);
        this.api.updateAlertStatus({id: this.idx[row_id], status: 'SUPPRESSED' })
    }   
  };

  updateAlertsStatusCleared = (status) => {
    for(var i in this.props.selectedRows.data) {
        let row_id = this.props.selectedRows.data[i].dataIndex
        // console.log(`Alert Row ${row_id} / ${this.idx[row_id]}`);
        this.api.updateAlertStatus({id: this.idx[row_id], status: 'CLEARED' })
    }   
  };

  render() {
    return (
      <div className={"custom-toolbar-select"}>
        <Tooltip title={"Change Status to CLEARED"}>
          <IconButton className={this.classes.iconButton} onClick={this.updateAlertsStatusCleared}>
            <Archive className={this.classes.icon} />
          </IconButton>
        </Tooltip>
        <Tooltip title={"Change Status to SUPPRESSED"}>
          <IconButton className={this.classes.iconButton} onClick={this.updateAlertsStatusSuppressed}>
            <Block className={this.classes.icon} />
          </IconButton>
        </Tooltip>
      </div>
    );
  }
}

export default withStyles(styles)(CustomToolbarSelect);
