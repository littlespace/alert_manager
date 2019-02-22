import React from 'react';
import { withStyles } from "@material-ui/core/styles";

import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import CardHeader from '@material-ui/core/CardHeader';
import CardMedia from '@material-ui/core/CardMedia';
import CardActions from '@material-ui/core/CardActions';
import Button from '@material-ui/core/Button';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';

const styles = theme => ({
    card: {
        width: "100%",
    },
    root: {
        flexGrow: 1,
      },
})


class SuppressionRuleItem extends React.Component {

    constructor(props){
      super(props);
      this.classes = this.props.classes;
      this.state = {}
    }
    

    render() {
      const subHeader = `Created by  ${this.props.data.Creator}`;

      return (

            <Grid item xs="5">
                <Card>
                    <CardHeader
                        title={this.props.data.Name}
                        subheader={subHeader} >
                    </CardHeader>
                    <CardContent>
                        {this.props.data.Reason}
                    </CardContent>
                    <CardActions>
                        <Button size="small" color="primary">
                            Delete
                        </Button>
                    </CardActions>
                </Card >
            </Grid>
      );
    }
  }
  
  export default withStyles(styles)(SuppressionRuleItem);