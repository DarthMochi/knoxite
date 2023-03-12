import React from "react";
import { Table, Card, Container } from 'react-bootstrap';
import 'react-rangeslider/lib/index.css';
import CodeBlock from './CodeBlock';
import { useTranslation } from "react-i18next";

const ClientInfo = ({client}) => {
    const { t } = useTranslation();
    const hostname = window.env.ADMIN_HOSTNAME;
    const port = window.env.ADMIN_UI_PORT;
    const protocol = window.env.SERVER_SCHEME;
    const repoUrl = protocol + "://" + client.AuthCode + "@" + hostname + (port !== "80" ? ":" + port + "" : "");

    return (
      <Container id="tutorial">
        <Card>
          <Card.Body>
            <h1>{t("tutorial.h1")}</h1>
            <h2>{t("tutorial.title")}</h2>
            <div>
              <ul>
                <li><a href="#intro">{t("tutorial.introduction.title")}</a></li>
                <li><a href="#init-repo">{t("tutorial.init_repo.title")}</a></li>
                <li><a href="#init-volume">{t("tutorial.init_volume.title")}</a></li>
                <li><a href="#listing-volumes">{t("tutorial.listing_volumes.title")}</a></li>
                <li><a href="#storing-data">{t("tutorial.storing_data.title")}</a></li>
              </ul>
            </div>
            <hr />
            <h2 id="intro">{t("tutorial.introduction.title")}</h2>
            <p>{t("tutorial.introduction.content")}</p>
            <Table className="table table-borderless">
              <thead></thead>
              <tbody>
                <tr>
                  <td>
                    <b>{t("client.auth_code")}</b>{": "}
                  </td>
                  <td>
                    {client.AuthCode}
                  </td>
                </tr>
                <tr>
                  <td>
                    <b>{t("knoxite_server_host")}</b>{": "}
                  </td>
                  <td>
                    {hostname}
                  </td>
                </tr>
                <tr>
                  <td>
                    <b>{t("knoxite_server_port")}</b>{": "}
                  </td>
                  <td>
                    {port}
                  </td>
                </tr>
                <tr>
                  <td>
                    <b>{t("knoxite_server_scheme")}</b>{": "}
                  </td>
                  <td>
                    {protocol}
                  </td>
                </tr>
              </tbody>
            </Table>
            <hr />
            <h2 id="init-repo">{t("tutorial.init_repo.title")}</h2>
            <p>
              {t("tutorial.init_repo.content1")}
            </p>
            <br />
            <CodeBlock isCopable={true}>
              knoxite -r {repoUrl} repo init
            </CodeBlock>
            <b>{t("tutorial.init_repo.content2")}</b>
            <hr />
            <h2 id="init-volume">{t("tutorial.init_volume.title")}</h2>
            <p>
              {t("tutorial.init_volume.content1")}
            </p>
            <CodeBlock isCopable={true}>
              knoxite -r {repoUrl} volume init "{t("tutorial.init_volume.volume_name")}" -d "{t("tutorial", "init_volume", "volume_description")}"
            </CodeBlock>
            <hr />
            <h2 id="listing-volumes">{t("tutorial.listing_volumes.title")}</h2>
            <p>
              {t("tutorial.listing_volumes.content1")}
            </p>
            <CodeBlock isCopable={true}>
              knoxite -r {repoUrl} volume list
            </CodeBlock>
            <p>
              {t("tutorial.listing_volumes.content2")}
            </p>
            <CodeBlock isCopable={false}>
              ID        Name                              Description
              ----------------------------------------------------------------------------------------------
              <br />
              66e03034  Volume name                       Volume description
            </CodeBlock>
            <p>
              {t("tutorial.listing_volumes.content3")}
            </p>
            <hr />
            <h2 id="storing-data">{t("tutorial.storing_data.title")}</h2>
            <p>
              {t("tutorial.storing_data.content1")}
            </p>
            <CodeBlock isCopable={true}>
              knoxite -r {repoUrl} store [{t("tutorial.storing_data.volume_id")}] {t("tutorial.storing_data.path_to_store")} -d "{t("tutorial.storing_data.snapshot_description")}"
            </CodeBlock>
            <p>
              {t("tutorial.storing_data.content2")}
            </p>
            <CodeBlock isCopable={false}>
              document.txt          5.69 MiB / 5.69 MiB [#########################################] 100.00%
              other.txt             4.17 MiB / 4.17 MiB [#########################################] 100.00%
              ...
              Snapshot cebc1213 created: 9 files, 8 dirs, 0 symlinks, 0 errors, 1.23 GiB Original Size, 1.23 GiB Storage Size
            </CodeBlock>
            <p>
              <span style={{fontSize: "30px"}}>
                &#128640;
              </span>
              <b>
                {t("tutorial.storing_data.content3")}
              </b>
              <span style={{fontSize: "30px"}}>
                &#10071;&#10071;&#10071;
              </span>
            </p>
            <p>
              {t("tutorial.getting_started")} <a href="https://knoxite.com/docs/getting-started/">Getting Started</a>
            </p>
          </Card.Body>
        </Card>
      </Container>
    );
};

export default ClientInfo;
