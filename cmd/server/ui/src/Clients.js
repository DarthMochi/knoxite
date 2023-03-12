import React, { useEffect, useRef } from "react";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrashAlt, faEdit, faPlus, faInfo, faCopy } from '@fortawesome/free-solid-svg-icons';
import { Table, Button, Card, } from "react-bootstrap";
import { useTranslation } from "react-i18next";
import { useAuth } from "./AuthProvider";
import { 
  sizeConversion, 
  deleteClientRequest,
  fetchClients,
  fetchStorageSize,
} from "./utils.js";
import { useNavigate } from "react-router-dom";

const Clients = ({setClients, setIsLoading, setError, clients, setClient, setStorageSize, setStorageSizeLabel, setAlert}) => {
  const { t } = useTranslation();
  const { token } = useAuth();
  const navigate = useNavigate();
  const isCalledRef = useRef(false);

  useEffect(() => {
    async function loadClients() {
      const cs = await fetchClients(navigate, token);
      if (cs !== null) {
        setClients(cs);
        const storageInfo = await fetchStorageSize(token);
        var [size, label] = sizeConversion(storageInfo, 0);
        setStorageSize(size);
        setStorageSizeLabel(label);
      } else {
        navigate("/login");
      }
    };
    if(!isCalledRef.current) {
      isCalledRef.current = true;
      loadClients();
      setIsLoading(false);
    }
  }, [setClients, navigate, setStorageSize, setStorageSizeLabel, setIsLoading, token]);

  return (
    <>
      <Card>
        <Card.Body>
          <Table hover size="sm">
            <TableHeader />
            <TableBody clientData={clients} token={token} setClients={setClients} setIsLoading={setIsLoading} setError={setError} setAlert={setAlert} setClient={setClient} />
          </Table>
          <Button variant="success" onClick={() => navigate("/admin/clients/new")}>
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

const TableBody = ({clientData, setClients, setClient, setIsLoading, setError, setAlert, token}) => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const deleteClient = async (client_id) => {
      if (window.confirm(t("delete_confirm_message")) === true) {
        setIsLoading(true);
        let response = await deleteClientRequest(token, client_id);
        if(response.ok) {
          setClients(clientData.filter(client => {
            return parseInt(client.ID) !== parseInt(client_id);
          }));
          setIsLoading(false);
        } else {
          setError(response.error);
          setAlert(response.error.message);
          setIsLoading(false);
        }
      }
  };

  const editForm = (index) => {
    var c = clientData[index];
    setClient(c);
    navigate("/admin/clients/" + c.ID + "/edit");
  }

  const clientInfo = (index) => {
    var c = clientData[index];
    setClient(c);
    navigate("/admin/clients/" + c.ID);
  }

  const clientElements = clientData.map((client, index) => {
    const UsedSpacePercentage = Math.round(100*(client.UsedSpace / client.Quota));
    const ProgressBarClassNames = quotaCSS(UsedSpacePercentage);
    const quota = client.Quota ? sizeConversion(client.Quota, 0) : 0;
    return (
      <tr key={index}>
        <td>{client.ID}</td>
        <td>{client.Name}</td>
        <td>
          <span id={"auth-code-" + client.ID}>
            {client.AuthCode}
          </span>
          <Button variant="light" onClick={() => {navigator.clipboard.writeText(document.getElementById("auth-code-" + client.ID).textContent)}}>
            <FontAwesomeIcon icon={faCopy} />
          </Button>
        </td>
        <td>{quota[0] + " " + quota[1]}</td>
        <td className="progress-cell">
          <div className="progress">
            <div className={ProgressBarClassNames} role="progressbar" style={{width: UsedSpacePercentage + "%"}} aria-valuenow={UsedSpacePercentage} aria-valuemin="0" aria-valuemax="100">
              {UsedSpacePercentage + "%"}
            </div>
          </div>
        </td>
        <td>
          <Button variant="danger" onClick={() => deleteClient(client.ID)}>
            <FontAwesomeIcon icon={faTrashAlt} />
          </Button>{' '}
          <Button variant="light" onClick={() => editForm(index)}>
            <FontAwesomeIcon icon={faEdit} />
          </Button>{' '}
          <Button variant="info" onClick={() => clientInfo(index)}>
            <FontAwesomeIcon icon={faInfo} />
          </Button>
        </td>
      </tr>
    );
  });
  return <tbody>{clientElements}</tbody>;
};

export default Clients;
