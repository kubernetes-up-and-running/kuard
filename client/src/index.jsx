import React from 'react';
import {render} from 'react-dom';
import Details from './details';
import Probe from './probe'
import Env from './env'

render(<Probe title="Liveness Check" path="/healthy" />, document.getElementById("liveness"));
render(<Probe title="Readiness Check" path="/ready" />, document.getElementById("readiness"));
render(<Env path="/env" />, document.getElementById("env"));

render(
  <Details title="File System">
      <a href="/fs/">Browse the root file system for this server.</a>
  </Details>, document.getElementById("fs"))
