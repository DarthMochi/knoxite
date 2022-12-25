//import { Children } from 'react';
import { Button } from 'react-bootstrap';
import { faCopy } from '@fortawesome/free-solid-svg-icons';
import 'react-rangeslider/lib/index.css';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

const CodeBlock = (props) => {
    const isCopable = props.isCopable;

    return (
      <div className='code-block'>
        { isCopable ? (
          <Button className='code-block-copy' variant="light" onClick={() => {navigator.clipboard.writeText(props.children.join(""))}}>
            <FontAwesomeIcon icon={faCopy} />
          </Button>
        ) : ("")}
        <pre>
          <code>
            {props.children}
          </code>
        </pre>
      </div>
    );
};

export default CodeBlock;
