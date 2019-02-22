
import React from 'react';
import { withStyles } from '@material-ui/core/styles';

import Menu from "../components/Menu"

import SuppressionRuleItem from '../components/SuppressionRule/SuppressionRuleItem'

import { AlertManagerApi } from '../library/AlertManagerApi';

import Grid from '@material-ui/core/Grid';


const AM = new AlertManagerApi()

const styles = theme => ({
    root: {
      flexGrow: 1,
      zIndex: 1,
      overflow: 'hidden',
      position: 'relative',
      display: 'flex',
      height: '100%',
    },
    gridroot: {
        flexWrap: "nowrap",
        margin: "10px"
    },
    content: {
      flexGrow: 1,
      backgroundColor: theme.palette.background.default,
      padding: theme.spacing.unit * 3,
      minWidth: 0, // So the Typography noWrap works
    },
});

class SuppressionRulesListView extends React.Component {

    constructor(props){
        super(props);
        this.classes = this.props.classes;
        this.state = {
            rules: []
        }
    };

    componentDidMount(){
        this.updateSuppRulesList()
    }

    updateSuppRulesList = () => {
        AM.getSuppressionRuleList()
          .then(data => this.setState({ rules: data }));
    }

    render() {
        let rules = this.state.rules
        return (
            <div>
                <div className={this.classes.root}>
                    <Menu />
                    <Grid 
                        className={this.classes.gridroot} 
                        container
                        spacing={16}
                        direction="column"
                        justify="flex-start"
                        alignItems="stretch"
                    >
                    { rules.map(r => {
                        return (
                        <SuppressionRuleItem
                            key={r.Id}
                            data={r}
                        >
                        </SuppressionRuleItem>
                        );
                    })}
                    </Grid>
                </div>
            </div>
        );
    }
}


export default withStyles(styles)(SuppressionRulesListView);
