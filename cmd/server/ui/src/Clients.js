import React, { useEffect, useRef, useState } from "react";
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
  convertSizeByStep,
  exponentSwitch,
} from "./utils.js";
import { useNavigate } from "react-router-dom";


const Clients = ({
  setClients, 
  setIsLoading, 
  setError, 
  clients, 
  setClient, 
  storageSizeLabel,
  setStorageSizeLabel, 
  setAlert, 
  setStorageSizeInBytes,
  storageSizeInBytes}) => {
  const { t } = useTranslation();
  const { token } = useAuth();
  const isCalledRef = useRef(false);
  const navigate = useNavigate();
  const [totalUsedSpace, setTotalUsedSpace] = useState(0);
  const [totalQuota, setTotalQuota] = useState(0);
  const [totalSpace, setTotalSpace] = useState(storageSizeInBytes);

  useEffect(() => {
    if(token === null) {
      setIsLoading(false);
      navigate("/admin/login");
    }
    setClient(null);

    if((clients || clients.length === 0) && !isCalledRef.current) {
      setTotalQuota(0);
      setTotalUsedSpace(0);
      setIsLoading(true);

      fetchClients(token).then((cs) => {
        if (cs !== null) {
          setClients(cs);
        }
      });
    }
      
    if(!isCalledRef.current) {
      fetchStorageSize(token).then((storageInfo) => {
        setStorageSizeInBytes(storageInfo);
        setTotalSpace(storageInfo);
        var label = sizeConversion(storageInfo, 0)[1];
        setStorageSizeLabel(label);

        var [tusp, tquo, tsp] = [0, 0, storageInfo];

        if(clients.length > 0) {
          clients.forEach((c, _index) => {
            tusp += c.UsedSpace;
            tquo += c.Quota;
            tsp += c.Quota;
          });
        }
        setTotalQuota(tquo);
        setTotalUsedSpace(tusp);
        setTotalSpace(tsp);
        isCalledRef.current = true;
        setIsLoading(false);
      });
    }
  }, [
    totalSpace,
    totalQuota,
    storageSizeInBytes,
    setClients, 
    setClient,
    clients, 
    setStorageSizeLabel, 
    setIsLoading, 
    setTotalUsedSpace,
    setTotalQuota,
    setTotalSpace,
    setStorageSizeInBytes,
    token,
    isCalledRef,
    navigate,
  ]);

  return (
    <>
      <Card>
        <Card.Body>
          <Table hover size="sm">
            <TableHeader />
            <TableBody 
              clientData={clients} 
              token={token} 
              setClients={setClients} 
              setIsLoading={setIsLoading} 
              setError={setError} 
              setAlert={setAlert} 
              setClient={setClient}
              setStorageSizeInBytes={setStorageSizeInBytes}
              setStorageSizeLabel={setStorageSizeLabel} />
            <TableFooter
              storageSizeLabel={storageSizeLabel}
              storageSpace={totalSpace}
              totalUsedSpace={totalUsedSpace}
              totalQuota={totalQuota} />
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

const TableBody = ({
  clientData, 
  setClients, 
  setClient, 
  setIsLoading, 
  setError, 
  setAlert, 
  token, 
  setStorageSizeInBytes,
  setStorageSizeLabel}) => {
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
          fetchStorageSize(token).then((storageInfo) => {
            setStorageSizeInBytes(storageInfo);
            var label = sizeConversion(storageInfo, 0)[1];
            setStorageSizeLabel(label);
            setIsLoading(false);
          });
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
          <Button variant="warning" onClick={() => {navigator.clipboard.writeText(document.getElementById("auth-code-" + client.ID).textContent)}}>
            <FontAwesomeIcon icon={faCopy} />
          </Button>{' '}
          <span id={"auth-code-" + client.ID}>
            {client.AuthCode}
          </span>
        </td>
        <td>{quota[0] + " " + quota[1]}</td>
        <td className="progress-cell">
          <div className="progress">
            <span className="progress-used-percentage">
              {UsedSpacePercentage + "%"}
            </span>
            <div className={ProgressBarClassNames} role="progressbar" style={{width: UsedSpacePercentage + "%"}} aria-valuenow={UsedSpacePercentage} aria-valuemin="0" aria-valuemax="100"></div>
          </div>
        </td>
        <td>
          <Button variant="danger" onClick={() => deleteClient(client.ID)}>
            <FontAwesomeIcon icon={faTrashAlt} />
          </Button>{' '}
          <Button variant="secondary" onClick={() => editForm(index)}>
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

const TableFooter =({
  totalQuota,
  totalUsedSpace,
  storageSpace,
  storageSizeLabel
  }) => {
  const UsedSpacePercentage = Math.round(100*(totalUsedSpace / storageSpace));
  const QuotaPercentage = Math.round(100*(totalQuota / storageSpace));
  const storageSize = convertSizeByStep(storageSpace, exponentSwitch(storageSizeLabel));
  const quotaSize = convertSizeByStep(totalQuota, exponentSwitch(storageSizeLabel));
  const usedSpaceSize = convertSizeByStep(totalUsedSpace, exponentSwitch(storageSizeLabel));
  const usedSpacePGClassNames = quotaCSS(UsedSpacePercentage);
  const quotaPGClassNames = quotaCSS(QuotaPercentage);
  return (
    <tfoot>
      <tr>
        <th>Total Given:</th>
        <th className="progress-cell" colSpan={2}>
          <div className="progress">
            <span className="progress-used-percentage">
              {quotaSize + " / " + storageSize + " " + storageSizeLabel}
            </span>
            <div className={quotaPGClassNames} role="progressbar" style={{width: QuotaPercentage + "%"}} aria-valuenow={UsedSpacePercentage} aria-valuemin="0" aria-valuemax="100"></div>
          </div>
        </th>
        <th>Total Used:</th>
        <th className="progress-cell">
          <div className="progress">
            <span className="progress-used-percentage">
              {usedSpaceSize + " / " + storageSize + " " + storageSizeLabel}
            </span>
            <div className={usedSpacePGClassNames} role="progressbar" style={{width: UsedSpacePercentage + "%"}} aria-valuenow={UsedSpacePercentage} aria-valuemin="0" aria-valuemax="100"></div>
          </div>
        </th>
        <th></th>
      </tr>
    </tfoot>
  )
};

export default Clients;
