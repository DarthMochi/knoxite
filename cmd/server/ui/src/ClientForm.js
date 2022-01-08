import React, { useState } from 'react';
import { Button, Card, Form, Container } from 'react-bootstrap';
import { fetchData } from './utils';

const ClientForm = (props) => {
    const isUpdate = false;
    const createClient = props.createClient;
    const updateClient = props.updateClient;
    const reset = props.reset;

    const [clientname, setClientName] = useState(props.client ? props.client.Name : '');

    const handleSubmit = (event) => {
        event.preventDefault();
        if (props.client) {
            console.log("Found client: ", props.client);
            updateClient(clientname);
        } else {
            createClient(clientname);
        }
        setClientName('');
        reset();
    };

    return (
        <Container>
            <Card>
                <Card.Body>
                    <Form onSubmit={handleSubmit} onReset={reset} >
                        <Form.Group className="mb-3" controlId="formBasicName">
                            <Form.Label>Client name</Form.Label>
                            <Form.Control type="text" placeholder="Enter client" value={clientname} onChange={(e) => setClientName(e.target.value)} />
                        </Form.Group>
                        <Button variant="light" type="submit">
                            Submit
                        </Button>{' '}
                        <Button variant="light" type="reset">
                            Cancel
                        </Button>
                    </Form>
                </Card.Body>
            </Card>
        </Container>
    );
};

export default ClientForm;