import React from 'react';
import { withRouter } from 'react-router-dom'
import { withStyles } from "@material-ui/core/styles";

import Grid from '@material-ui/core/Grid';

import { 
    timeConverter, 
    secondsToHms 
} from '../../library/utils';

const styles = theme => ({
    card: {
        width: "100%",
    },
    root: {
        flexGrow: 1,
    },
    historyItem: {
        fontSize: "0.75rem",
        letterSpacing: "0.01071em",
        textAlign: "center",
        verticalAlign: "middle",
        top: "50%",
        border: 1,
        borderBottom: "1px solid rgba(224, 224, 224, 1)",
        // lineeight: 90;
    },
    historyCell: {
        lineHeight: "40px",
        height: "40px",
        textAlign: "left",
        paddingLeft: 10
    }
})



class historyItem extends React.Component {

    constructor(props, context){
      super(props, context);
      this.classes = this.props.classes;
      this.state = {}
    }
    
    render() {
      return (
            <Grid container item xs={12} className={this.classes.historyItem}>
                <Grid item xs={12}  sm={2} className={this.classes.historyCell}>{timeConverter(this.props.data.Timestamp)}</Grid>
                <Grid item xs={12}  sm={8} className={this.classes.historyCell}>{this.props.data.Event}</Grid>
                <Grid item xs={12}  sm={2} className={this.classes.historyCell}>{secondsToHms(this.props.data.Timestamp)}</Grid>
            </Grid>
      );
    }
  }
  
export default withRouter(withStyles(styles)(historyItem))