import React from 'react';
import { withRouter } from 'react-router-dom'
import { withStyles } from "@material-ui/core/styles";
import classNames from 'classnames';

import Button from '@material-ui/core/Button';
import Grid from '@material-ui/core/Grid';
import Avatar from '@material-ui/core/Avatar';
import Tooltip from '@material-ui/core/Tooltip';

import AssignmentTurnedInIcon from '@material-ui/icons/AssignmentTurnedIn';

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
    alertItem: {
        fontSize: "0.75rem",
        letterSpacing: "0.01071em",
        textAlign: "center",
        verticalAlign: "middle",
        top: "50%",
        border: 1,
        borderBottom: "1px solid rgba(224, 224, 224, 1)",
        cursor: "pointer"
        // lineeight: 90;
    },
    alertCell: {
        lineHeight: "40px",
        height: "40px",
    },
    alertItemName: {
        textAlign: "left"
    },
    alertItemTimes: {
        textAlign: "left",
        lineHeight: "20px"
    },
    alertItemTimeItem: {
        lineHeight: "20px",
        height: "20px"
    },
    button: {
        fontSize: "0.6rem",
        padding: '2px 2px',
        minHeight: 10,
        marginRight: 2,
        marginLeft: 2,
    },
    ownerBadge: {
        // padding: '0px 2px',
        color: '#fff',
        // backgroundColor: "#32CD32",
        backgroundColor: "#424242",
        height: 26,
        width: 26,
        marginTop: 7
        // marginTop: "8px"
    },
    ownerBadgeIcon: {
        height: 15,
        width: 15,
    },
    alertWarn: {
        backgroundColor: '#FFF3E0'
    },
    alertCritical: {
        backgroundColor: '#ffebee'  
    },
    alertInfo: {
        backgroundColor: '#E3F2FD'
    },
})

const alert_mapping = {
    CRITICAL: 'alertCritical',
    WARN: 'alertWarn',
    INFO: 'alertInfo'
}


class AlertItem extends React.Component {
    constructor(props, context){
      super(props, context);
      this.classes = this.props.classes;
      this.state = {}
    }
    
    redirectToAlert = () => {
        this.props.history.push(`/alert/${this.props.data.Id}`)
    }

    render() {

      return (
        
            <Grid container item 
                xs={12}
                onClick={this.redirectToAlert}
                className={classNames(this.classes.alertItem,this.classes[alert_mapping[this.props.data.Severity]])}>
                <Grid container item xs={4} sm={1} md={1} className={this.classes.alertCell}>
                    <Grid item xs={9} sm={8}>
                        <Button variant="contained" color="primary" className={this.classes.button}>
                            {(this.props.data.Status === 'SUPPRESSED') ? "SUPPR" : this.props.data.Status}
                        </Button>
                    </Grid>
                    <Grid item xs={3} sm={4} >
                        { (this.props.data.Owner !== "") ? 
                        <Tooltip title={this.props.data.Owner} placement="bottom">
                            <Avatar className={this.classes.ownerBadge}>
                                <AssignmentTurnedInIcon className={this.classes.ownerBadgeIcon}/>
                            </Avatar>
                        </Tooltip> : "" }
                    </Grid>
                </Grid>
                <Grid item xs={8} sm={3} md={4} className={classNames(this.classes.alertCell,this.classes.alertItemName)}>{this.props.data.Name}</Grid>
                <Grid container item xs={6} sm={2} md={2} className={classNames(this.classes.alertCell,this.classes.alertItemTimes)}>
                    <Grid item xs={12}>Site: {(this.props.data.Site !== "") ? this.props.data.Site : "undefined" }</Grid>
                    <Grid item xs={12}>Device: {(this.props.data.Device!== "") ? this.props.data.Device : "undefined"}</Grid>
                </Grid>
                <Grid item xs={6} sm={1} className={this.classes.alertCell}>{this.props.data.Scope}</Grid>
                <Grid item xs={6} sm={3} md={2} className={this.classes.alertCell}>{this.props.data.Source}</Grid>
                <Grid container item xs={6} sm={2} className={classNames(this.classes.alertCell,this.classes.alertItemTimes)}>
                    <Grid item xs={12} className={this.classes.alertItemTimeItem}>Start: {timeConverter(this.props.data.start_time)}</Grid>
                    <Grid item xs={12} className={this.classes.alertItemTimeItem}>Last Active: {secondsToHms(this.props.data.last_active)}</Grid>
                </Grid>
            </Grid>
        
      );
    }
  }
  
export default withRouter(withStyles(styles)(AlertItem))