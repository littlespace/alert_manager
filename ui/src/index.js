import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';
import registerServiceWorker from './registerServiceWorker';
// import Dashboard from './Dashboard';
import App from './components/App';
import { BrowserRouter } from 'react-router-dom';

// ReactDOM.render(<Dashboard />, document.getElementById('root'));
// ReactDOM.render(<App />, document.getElementById('root'));
ReactDOM.render((
    <BrowserRouter>
      <App />
    </BrowserRouter>
  ), document.getElementById('root'))

registerServiceWorker();
