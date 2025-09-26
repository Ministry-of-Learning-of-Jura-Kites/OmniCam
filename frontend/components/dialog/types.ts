export type InputTypes = "text" | "number" | "textarea" | "file";

export type InputTypeMap = {
  text: string;
  textarea: string;
  file: File;
  number: number;
};

export type FieldOption = {
  type: InputTypes;
  required?: boolean;
};
