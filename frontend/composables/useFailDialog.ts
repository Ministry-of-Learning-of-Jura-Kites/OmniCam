export const useFailDialog = () => {
  const open = useState<boolean>("failDialogOpen", () => false);
  const message = useState<string>("failDialogMessage", () => "");

  function showFailDialog(msg: string) {
    message.value = msg;
    open.value = true;
  }

  function closeFailDialog() {
    open.value = false;
  }

  return { open, message, showFailDialog, closeFailDialog };
};
