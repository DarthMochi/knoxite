import i18next from "i18next";
import { initReactI18next } from "react-i18next";

const resources = {
  en: {
    translation: {
      knoxite_server_host: "Hostname",
      knoxite_server_port: "Port",
      knoxite_server_scheme: "Protocol",
      login: {
        title: "Login",
        username_placeholder: "Username",
        password_placeholder: "Password",
      },
      delete_confirm_message: "Are you sure?",
      admin_name: "Admin username",
      admin_password: "Admin password",
      submit: "Submit",
      cancel: "Cancel",
      loading: "Loading",
      client: {
        new_button: "Create new client",
        id: "ID",
        name: "Client name",
        auth_code: "AuthCode",
        quota: "Quota",
        used_space: "Used space",
        repo_url: "Repository URL",
        password: "Repository password",
        name_placeholder: "Enter client name",
        created_success: "Client created successfully",
        updated_success: "Client updated successfully",
        deleted_success: "Client created successfully",
      },
      volumes: {
        more: "Volumes",
        id: "ID",
        name: "Name",
        description: "Description",
      },
      snapshots: {
        more: "Snapshots",
        id: "ID",
        created_at: "Created at",
        size: "Size",
        storage_size: "Storage size",
        description: "Description",
      },
      files: {
        all: "All Files",
        more: "Files",
        mode: "Mode",
        group: "Group",
        user: "User",
        size: "Size",
        modified: "Modified",
        name: "Name",
        no_files_found: "No files in this folder."
      },
      tutorial: {
        h1: "Quick Guide to get knoxited",
        title: "Content",
        introduction: {
          title: "Introduction",
          content: "This is a short tutorial on how to get your client " +
          "setup, to backup data on this knoxite server. The commands include already the correct " +
          "username, which is the authentication code, for you repository as well as the correct protocol, " +
          "host and port. The basic configuration values are the following: ",
        },
        init_repo: {
          title: "Step 1: Initializing a Repository",
          content1: "First of all we need to initialize a repository on this server. " +
          "You will be asked for a password, which will be used to encrypt your data. This you will need to " +
          "repeat twice.",
          content2: "Be warned: If you lose this password, you won't be able " +
          "to access any of your data.",
        },
        init_volume: {
          title: "Step 2: Initializing a Volume",
          content1: "Each repository can contain several volumes, which store " +
          "our data organized in snapshots. You will need to specify the volume name after init and a " +
          "description with the flag \"-d\".",
          volume_name: "Enter here a custom volume name",
          volume_description: "Enter here a custom volume description",
        },
        listing_volumes: {
          title: "Step 3: Listing Volumes",
          content1: "To store data, you will need to the volume id, where to store " +
          "the backups. You can get a list of all volumes stored in this repository with the following command:",
          content2: "This will display a table on the console which looks like this: ",
          content3: "The volume ID can be found in the most left column (ID).",
        },
        storing_data: {
          title: "Step 4: Storing data",
          content1: "Run the following command to create a new snapshot and store your " +
          "home directory in the newly created volume:",
          volume_id: "Enter the volume id you got from the chapter before here (remove brackets)",
          path_to_store: "Enter here your path to store.",
          snapshot_description: "Enter a comment for your snapshot here",
          content2: "This will create a snapshot and will store your data on the knoxite-server." +
          "The output will look something like this:",
          content3: "CONGRATULATIONS: You have successfully stored data on your knoxite-server",
        },
        getting_started: "For more information how to get started, visit:",
      },
    },
  },
};
i18next
  .use(initReactI18next)
  .init({
    resources: resources,
    lng: "en",
    interpolation: {
      escapeValue: false,
    },
    debug: window.env.NODE_ENV === 'development',
  });

export default i18next;
