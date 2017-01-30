import React from 'react';
import marked from 'marked';

export default class MarkdownElement extends React.Component {
  constructor(props) {
    super(props);

    let renderer = new marked.Renderer()
    renderer.table = function(header, body) {
      return '<table class="table table-condensed table-bordered">\n'
        + '<thead>\n'
        + header
        + '</thead>\n'
        + '<tbody>\n'
        + body
        + '</tbody>\n'
        + '</table>\n';
    };

    marked.setOptions({
      renderer: renderer,
      gfm: true,
      tables: true,
      breaks: false,
      pedantic: false,
      sanitize: true,
      smartLists: true,
      smartypants: false
    });
  }
  render() {
    const { text } = this.props,
    html = marked(text || '');

    return (
      <div>
        <div dangerouslySetInnerHTML={{__html: html}} />
      </div>
    );
  }
}

MarkdownElement.propTypes = {
  text: React.PropTypes.string.isRequired
};

MarkdownElement.defaultProps = {
  text: ''
};
