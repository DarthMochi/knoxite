import React, { useCallback, useEffect, useRef, useState } from "react";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrashAlt, faEdit, faPlus, faInfo, faCopy } from '@fortawesome/free-solid-svg-icons';
import { Table, Button, Card, } from "react-bootstrap";
import { useTranslation } from "react-i18next";
import { useAuth } from "./AuthProvider";
import { 
  sizeConversion, 
  deleteClientRequest,
  fetchClients,
  convertSizeByStep,
  exponentSwitch,
  fetchTotalQuota,
  fetchUsedSpace,
  fetchStorageSpace,
} from "./utils.js";
import { useNavigate } from "react-router-dom";


const Clients = ({
  setClients, 
  loadingHandler, 
  setError, 
  clients, 
  setClient, 
  setAlert,
  totalQuota, 
  setTotalQuota,
  usedSpace, 
  setUsedSpace,
  storageSpace, 
  setStorageSpace
  }) => {
  const { t } = useTranslation();
  const { token } = useAuth();
  const clientsRef = useRef(false);
  const totalQuotaRef = useRef(false);
  const usedSpaceRef = useRef(false);
  const storageSpaceRef = useRef(false);
  const navigate = useNavigate();

  const loadClients = useCallback(() => {
    if(!clientsRef.current) {
      clientsRef.current = true;
      loadingHandler("load_clients", "push");
      fetchClients(token).then((cs) => {
        if (cs === null) {
          cs = []
        }
        setClients(cs);
        loadingHandler("load_clients", "pop");
      });
    }
  }, [
    token,
    setClients,
    loadingHandler
  ]);

  const loadTotalQuota = useCallback(() => {
    if(!totalQuotaRef.current) {
      totalQuotaRef.current = true;
      loadingHandler("total_quota", "push");
      fetchTotalQuota(token).then((tq) => {
        if(tq != null) {
          setTotalQuota(tq);
        } else {
          setTotalQuota(0);
        }
        loadingHandler("total_quota", "pop");
      });
    }
  }, [
    token,
    setTotalQuota,
    loadingHandler,
  ]);

  const loadUsedSpace = useCallback(() => {
    if(!usedSpaceRef.current) {
      usedSpaceRef.current = true;
      loadingHandler("used_space", "push");
      fetchUsedSpace(token).then((us) => {
        if(us !== null) {
          setUsedSpace(us);
        } else {
          setUsedSpace(0);
        }
        loadingHandler("used_space", "pop");
      });
    }
  }, [
    token,
    setUsedSpace,
    loadingHandler,
  ])

  const loadStorageSpace = useCallback(() => {
    if(!storageSpaceRef.current) {
      storageSpaceRef.current = true;
      loadingHandler("storage_space", "push");
      fetchStorageSpace(token).then((ss) => {
        if(ss !== null) {
          setStorageSpace(ss);
        } else {
          setStorageSpace(0);
        }
        loadingHandler("storage_space", "pop");
      });
    }
  }, [
    token,
    setStorageSpace,
    loadingHandler,
  ]);

  useEffect(() => {
    if(token === null) {
      navigate("/admin/login");
    } else {
      loadClients();
      loadTotalQuota();
      loadUsedSpace();
      loadStorageSpace();
    }
  }, [
    token,
    loadClients,
    loadTotalQuota,
    loadUsedSpace,
    loadStorageSpace,
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
              loadingHandler={loadingHandler} 
              setError={setError} 
              setAlert={setAlert} 
              setClient={setClient}
              totalQuota={totalQuota}
              setTotalQuota={setTotalQuota}
              usedSpace={usedSpace}
              setUsedSpace={setUsedSpace} />
            <TableFooter 
              totalQuota={totalQuota}
              usedSpace={usedSpace}
              storageSpace={storageSpace} />
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
  loadingHandler, 
  setError, 
  setAlert, 
  totalQuota, 
  setTotalQuota,
  usedSpace, 
  setUsedSpace,
  token}) => {
  const navigate = useNavigate();
  const { t } = useTranslation();

  const deleteClient = async (client_id, index) => {
    if (window.confirm(t("delete_confirm_message")) === true) {
      loadingHandler("delete_client", "push");
      var c = clientData[index];
      var total_quota = totalQuota - c.Quota;
      var used_space = usedSpace - c.UsedSpace;
      let response = await deleteClientRequest(token, client_id).finally(() => {
        loadingHandler("delete_client", "pop");
      });

      if(response.ok) {
        setClients(clientData.filter(client => {
          return parseInt(client.ID) !== parseInt(client_id);
        }));
        setClient(null);
        setTotalQuota(total_quota);
        setUsedSpace(used_space);
      } else {
        setError(response.error);
        setAlert(response.error.message);
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
          <Button variant="danger" onClick={() => deleteClient(client.ID, index)}>
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
  usedSpace,
  storageSpace}) => {
  const [storageSizeLabel, setStorageSizeLabel] = useState("");
  const [UsedSpacePercentage, setUsedSpacePercentage] = useState(0);
  const [QuotaPercentage, setQuotaPercentage] = useState(0);
  const [storageSize, setStorageSize] = useState(0);
  const [quotaSize, setQuotaSize] = useState(0);
  const [usedSpaceSize, setUsedSpaceSize] = useState(0);
  const [usedSpacePGClassNames, setUsedSpacePGClassNames] = useState("");
  const [quotaPGClassNames, setQuotaPGClassNames] = useState("");
  // const ref = useRef({totalQuota, usedSpace, storageSpace, clients});

  useEffect(() => {
    console.log(totalQuota);
    console.log(usedSpace);
    console.log(storageSpace);
    if(totalQuota !== null && usedSpace !== null && storageSpace !== null) {
      console.log("data loaded");
      var label = sizeConversion(storageSpace, 0)[1];
      var size = convertSizeByStep(storageSpace, exponentSwitch(label))
      var usedSpacePercentage = Math.round(100*(usedSpace / storageSpace));
      var quotaPercentage = Math.round(100*(totalQuota / storageSpace));

      setStorageSize(size);
      setStorageSizeLabel(label);
      setUsedSpacePercentage(usedSpacePercentage);
      setQuotaPercentage(quotaPercentage);
      setQuotaSize(convertSizeByStep(totalQuota, exponentSwitch(label)));
      setUsedSpaceSize(convertSizeByStep(usedSpace, exponentSwitch(label)));
      setUsedSpacePGClassNames(quotaCSS(usedSpacePercentage));
      setQuotaPGClassNames(quotaCSS(quotaPercentage));
    }
  }, [
    totalQuota,
    usedSpace,
    storageSpace,
    setStorageSize,
    setStorageSizeLabel,
    setUsedSpacePercentage,
    setQuotaPercentage,
    setQuotaSize,
    setUsedSpaceSize,
    setUsedSpacePGClassNames,
    setQuotaPGClassNames,
  ]);

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
