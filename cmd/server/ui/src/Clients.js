import React, { useEffect, useRef } from "react";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrashAlt, faEdit, faPlus, faInfo } from '@fortawesome/free-solid-svg-icons';
import { Table, Button, Card, } from "react-bootstrap";
import { useTranslation } from "react-i18next";
import { useAuth } from "./AuthProvider";
import { 
  sizeConversion, 
  deleteClientRequest,
  fetchClients,
  fetchStorageSize,
} from "./utils.js";

const Clients = (props) => {
  const { t } = useTranslation();
  const { token } = useAuth();
  const isCalledRef = useRef(false);

  useEffect(() => {
    async function loadClients() {
      props.setIsLoading(true);
      props.setError(false);
      const clients = await fetchClients(token);
      props.setClients(clients);
      const storageInfo = await fetchStorageSize(token);
      var [size, label] = sizeConversion(storageInfo, 0);
      props.setStorageSize(size);
      props.setStorageSizeLabel(label);
      props.setIsLoading(false);
    };
    if(!isCalledRef.current) {
      isCalledRef.current = true;
      loadClients();
    }
  }, [props, token]);

  const deleteClient = async (index, client_id) => {
    if (window.confirm(t("delete_confirm_message")) === true) {
      props.setIsLoading(true);
      let response = await deleteClientRequest(token, client_id);
      if(response.ok) {
        props.setClients(props.clients.filter((_, i) => {
          return i !== index;
        }));
        props.setIsLoading(false);
      } else {
        props.setError(response.error);
        props.setAlert(response.error.message);
        props.setIsLoading(false);
      }
    }
  };

  return (
    <>
      <Card>
        <Card.Body>
          <Table hover size="sm">
            <TableHeader />
            <TableBody clientData={props.clients} deleteClient={deleteClient} token={token} />
          </Table>
          <Button variant="success" href="/admin/clients/new">
            <FontAwesomeIcon icon={faPlus} /> {t("client.new_button")}
          </Button>
        </Card.Body>
      </Card>
    </>
  );
};

const quotaCSS = (quota) => {
    var className = "progress-bar";
    if(quota < 25) {
        className += " bg-success";
    } else if(quota < 50) {
        className += " bg-info";
    } else if(quota < 75) {
        className += " bg-warning";
    } else {
        className += " bg-danger";
    }

    return className;
}

const TableHeader = () => {
    const { t } = useTranslation();
    return (
        <thead>
            <tr>
                <th>{t("client.id")}</th>
                <th>{t("client.name")}</th>
                <th>{t("client.auth_code")}</th>
                <th>{t("client.quota")}</th>
                <th width="25%">{t("client.used_space")}</th>
                <th></th>
            </tr>
        </thead>
    );
};

const TableBody = (props) => {
    const clients = props.clientData.map((client, index) => {
      const UsedSpacePercentage = Math.round(100*(client.UsedSpace / client.Quota));
      const ProgressBarClassNames = quotaCSS(UsedSpacePercentage);
      const quota = sizeConversion(client.Quota, 0);
      return (
        <tr key={index}>
          <td>{client.ID}</td>
          <td>{client.Name}</td>
          <td>{client.AuthCode}</td>
          <td>{quota[0] + " " + quota[1]}</td>
          <td className="progress-cell">
            <div className="progress">
              <div className={ProgressBarClassNames} role="progressbar" style={{width: UsedSpacePercentage + "%"}} aria-valuenow={UsedSpacePercentage} aria-valuemin="0" aria-valuemax="100"></div>
              <span className="progress-bar-label">
                {UsedSpacePercentage + "%"}
              </span>
            </div>
          </td>
          <td>
            <Button variant="danger" onClick={() => props.deleteClient(index, client.ID)}>
              <FontAwesomeIcon icon={faTrashAlt} />
            </Button>{' '}
            <Button variant="light" href={"/admin/clients/" + client.Name + "/edit"}>
              <FontAwesomeIcon icon={faEdit} />
            </Button>{' '}
            <Button varian="light" href={"/admin/clients/" + client.Name}>
              <FontAwesomeIcon icon={faInfo} />
            </Button>
          </td>
        </tr>
      );
    });
  return <tbody>{clients}</tbody>;
};

export default Clients;
