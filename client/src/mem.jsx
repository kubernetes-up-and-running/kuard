import React from 'react';
import numeral from 'numeral';
import fetchError from './fetcherror';

export default class Mem extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      memStats: {}
    };
  }

  loadState() {
    fetch(this.props.apiPath)
    .then(response => response.json())
    .then(response => this.setState(state => {
        state.memStats = response.memStats;
        delete state.memStats.PauseNs;
        delete state.memStats.PauseEnd;
        delete state.memStats.BySize;
        return state
    }));
  }

  componentDidMount() {
    this.loadState()
    this.timer = setInterval(this.loadState.bind(this), 1000);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  alloc(e, n) {
    e.preventDefault();
    fetch(this.props.apiPath+"/alloc?size=" + n, {
      method: "POST",
    })
    .then(fetchError)
    .catch(err => this.context.reportConnError());
  }

  clear(e) {
    e.preventDefault();
    fetch(this.props.apiPath+"/clear", {
      method: "POST",
    })
    .then(fetchError)
    .catch(err => this.context.reportConnError());
  }

  addRow(rows, k, v) {
      rows.push(
        <tr key={k}>
          <td><samp>{k}</samp></td>
          <td><samp>{v}</samp></td>
        </tr>
      )  }

  render () {
    let rows = [];

    let m = this.state.memStats;
    let format = '0.00 ib';
    let tot = 0;
    this.addRow(rows, "HeapAlloc", numeral(m.HeapAlloc).format(format));
    tot += m.HeapAlloc;
    this.addRow(rows, "HeapIdle - HeapReleased", numeral(m.HeapIdle - m.HeapReleased).format(format));
    tot += m.HeapIdle - m.HeapReleased;
    this.addRow(rows, "StackInuse", numeral(m.StackInuse).format(format));
    tot += m.StackInuse
    this.addRow(rows, "Total", numeral(tot).format(format))

    return (
      <div>
        <table className="table table-condensed table-bordered">
          <thead>
            <tr>
              <th>Key</th><th>Value</th>
            </tr>
          </thead>
          <tbody>
            {rows}
          </tbody>
        </table>
        <div>
          <b>Allocate/Clear memory to force OOM</b><br/>
          <a className="failn" onClick={e => this.alloc(e, 500*1024*1024)} href="#">Allocate 500 MiB</a><br/>
          <a className="failn" onClick={e => this.clear(e)} href="#">Clear</a>
        </div>
      </div>
    )
  }
}

Mem.propTypes =  {
  apiPath: React.PropTypes.string.isRequired,
}

Mem.contextTypes = {
  reportConnError: React.PropTypes.func
};
