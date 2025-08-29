export interface FieldConfig {
  key: string;
  type: InputTypes;
}

type InputTypes = "text" | "number" | "textarea" | "file";
