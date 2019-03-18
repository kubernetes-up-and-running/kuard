import React from 'react';
import fetchError from './fetcherror';
import Markdown from './markdown'

const apiMD = ` This shows the status of a simple in memory queue.  This is
based heavily on https://github.com/kelseyhightower/memq.

The API is as follows with URLs being relative to \`<server addr>/<api-base>/memq/server\`.
See \`pkg/memq/types.go\` for the data structures returned.

| Method | Url | Desc
| --- | --- | ---
| \`GET\` | \`/stats\` | Get stats on all queues
| \`PUT\` | \`/queues/:queue\` | Create a queue
| \`DELETE\` | \`/queue/:queue\` | Delete a queue
| \`POST\` | \`/queue/:queue/drain\` | Discard all items in queue
| \`POST\` | \`/queue/:queue/enqueue\` | Add item to queue.  Body is plain text. Response is message object.
| \`POST\` | \`/queue/:queue/dequeue\` | Grab an item off the queue and return it. Returns a 204 "No Content" if queue is empty.
`

export default class MemQ extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      stats: {
        kind: "stats",
        queues: []
      }
    };
  }

  loadState() {
    fetch(this.props.serverPath+"/server/stats")
    .then(fetchError)
    .then(response => response.json())
    .then(response => this.setState(state => {
      response.queues.sort((a,b) => {
        if (a.name < b.name) return -1;
        if (a.name > b.name) return 1;
        return 0;
      })
      state.stats = response;
      return state
    }))
    .catch(err => this.context.reportConnError());
  }

  componentDidMount() {
    this.loadState()
    this.timer = setInterval(this.loadState.bind(this), 1000);
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  render () {
    let status = ""
    if (this.state.stats.queues.length == 0) {
      status = (<div> There are no queues created </div>)
    } else {
      let rows = [];
      for (let q of this.state.stats.queues) {
        rows.push(
          <tr key={q.name}>
            <td>{q.name}</td>
            <td>{q.depth}</td>
            <td>{q.enqueued}</td>
            <td>{q.dequeued}</td>
            <td>{q.drained}</td>
          </tr>
        )
      }
      status = (
        <table className="table table-condensed table-bordered">
          <thead>
            <tr>
              <th>Name</th>
              <th>Depth</th>
              <th>Enqueued</th>
              <th>Dequeued</th>
              <th>Drained</th>
            </tr>
          </thead>
          <tbody>
            {rows}
          </tbody>
        </table>
      )
    }

    return (
      <div>
        {status}
        <hr/>
        <Markdown text={apiMD}/>
      </div>
    )
  }
}

MemQ.propTypes =  {
  serverPath: React.PropTypes.string.isRequired
}

MemQ.contextTypes = {
  reportConnError: React.PropTypes.func
};
