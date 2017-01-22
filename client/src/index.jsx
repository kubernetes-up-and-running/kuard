import React from 'react';
import {render} from 'react-dom';
import Details from './details';
import Probe from './probe'

render(<Probe path="/foo" />, document.getElementById('test'));
