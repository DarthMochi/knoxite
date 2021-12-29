import React from "react";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faTrashAlt, faEdit } from '@fortawesome/free-solid-svg-icons'
import { Container, Table, Button } from "react-bootstrap";

const Clients = (props) => {
    const {clientData, deleteClient} = props

    return (
        <Container fluid>
            <Table hover size="sm">
                <TableHeader />
                <TableBody clientData={clientData} deleteClient={deleteClient}/>
            </Table>
        </Container>
    )
}

const TableHeader = () => {
    return (
        <thead>
            <tr>
                <th>ID</th>
                <th>Name</th>
                <th>AuthCode</th>
                <th></th>
            </tr>
        </thead>
    )
}

const TableBody = (props) => {
    const clients = props.clientData.map((client, index) => {
        return (
            <tr key={index}>
                <td>{client.ID}</td>
                <td>{client.Name}</td>
                <td>{client.AuthCode}</td>
                <td>
                    <Button variant="light" onClick={() => props.deleteClient(index)}>
                        <FontAwesomeIcon icon={faTrashAlt} />
                    </Button>{' '}
                    <Button variant="light" onClick={() => console.log("not implemented")}>
                        <FontAwesomeIcon icon={faEdit} />
                    </Button>
                </td>
            </tr>
        )
    })
    return <tbody>{clients}</tbody>
}

export default Clients;