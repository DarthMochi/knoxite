import { Alert } from "react-bootstrap";

const ErrorMessage = (props) => {
    if (props.message) {
        const variant = props.err ? 'danger' : 'success';
        return <Alert variant={variant}>{props.message}</Alert>
    }
    return <div></div>
}

export default ErrorMessage;