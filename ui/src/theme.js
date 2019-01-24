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
  });
