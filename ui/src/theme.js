import { createMuiTheme } from '@material-ui/core/styles';
import red from '@material-ui/core/colors/red';

export default createMuiTheme({
    palette: {
      primary: {
        main: '#424242',
      },
      secondary: red,
    },
    status: {
      danger: 'orange',
    },
    typography: {
      useNextVariants: true,
    },
    pageTitle:{
      height: "30px",
      lineHeight: "30px",
      paddingLeft: "15px",
      paddingTop: "10px"
    }
  });
