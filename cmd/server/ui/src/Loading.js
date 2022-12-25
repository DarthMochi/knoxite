import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faSpinner } from '@fortawesome/free-solid-svg-icons';
import { useTranslation } from "react-i18next";

const Loading = (props) => {
  const { t } = useTranslation();
  return (
    <div className="loading" style={{display: props.isLoading ? "flex" : "none"}}>
      <div className="loading-spinner">
        <FontAwesomeIcon size="lg" icon={faSpinner} className="spinner" />
        <br />
        <span>{t("loading")}</span>
      </div>
    </div>
  );
};

export default Loading;
