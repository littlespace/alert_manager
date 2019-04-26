   
   
   
import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';

import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';

import Typography from '@material-ui/core/Typography';


const styles = theme => ({
    legendAlertInfo: {
        margin: 10,
        backgroundColor: '#E3F2FD',
      },
      legendAlertWarn: {
        margin: 10,
        backgroundColor: '#FFF3E0',
      },
      legendAlertCritical: {
        margin: 10,
        backgroundColor: '#ffebee',
      },
})

// alertWarn: {
//     backgroundColor: '#FFF3E0'
// },
// alertCritical: {
//     backgroundColor: '#ffebee'  
// },
// alertInfo: {
//     backgroundColor: '#E3F2FD'
// },


const Text =  ({content}) => {
    return (
       <p dangerouslySetInnerHTML={{__html: content}}></p>
    );
};

class PageHelp extends React.Component {

    constructor(props, context) {
        super(props, context);
        this.classes = this.props.classes;
        this.state = {
            open: this.props.open
        }
    }

    componentWillReceiveProps(nextProps) {
        if (nextProps.open !== this.state.open) {
            this.setState({ open: nextProps.open });
        }
    };

    handleClose = () => {
        this.setState({ open: false });
    };

    render() {
        return (
            <div>
                <Dialog
                    onClose={this.handleClose}
                    aria-labelledby="customized-dialog-title"
                    open={this.state.open}
                    >
                    <DialogTitle id="customized-dialog-title" onClose={this.handleClose}>
                        {this.props.title}
                    </DialogTitle>
                    <DialogContent>
                    <Text 
                    //   className={this.classes.alertDescriptionText} 
                      content={this.props.description}
                    ></Text>
{/* 
                        <Typography gutterBottom>
                        {this.props.description}
                        </Typography> */}
                    </DialogContent>
                    <DialogContent>
                    <Typography variant="h6">
                        Alert Legend
                        </Typography>
                    <Text 
                      content="The color of the alert refects its severity:"
                    ></Text>
                    <div>
                    <Button size="small" disabled className={this.classes.legendAlertInfo}>
                        INFO
                    </Button>
                    <Button size="small" disabled className={this.classes.legendAlertWarn}>
                    WARNING
                    </Button>
                    <Button size="small" disabled className={this.classes.legendAlertCritical}>
                    CRITICAL
                    </Button>
                    </div>
                    </DialogContent>

                    <DialogActions>
                        <Button onClick={this.handleClose} color="primary">
                        Close
                        </Button>
                    </DialogActions>
                </Dialog>
            </div>
    )};
}
 
 export default withStyles(styles)(PageHelp);
 