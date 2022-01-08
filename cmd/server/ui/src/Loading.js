import { Container, Row, Col } from "react-bootstrap";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faSpinner } from '@fortawesome/free-solid-svg-icons'

const Loading = () => {
    return (
        <Container>
            <Row className="justify-content-md-center">
                <Col md="auto">
                    <FontAwesomeIcon icon={faSpinner} />
                    <p>Loading ...</p>
                </Col>
            </Row>
        </Container>
    );
};

export default Loading;