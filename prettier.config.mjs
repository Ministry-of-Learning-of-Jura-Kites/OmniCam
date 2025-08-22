import { postgresql } from "sql-formatter";

/** @type {import('prettier').Config} */
const prettierConfig = {
  plugins: ["prettier-plugin-embed", "prettier-plugin-sql"],
};

/** @type {import('prettier-plugin-embed').PrettierPluginEmbedOptions} */
const prettierPluginEmbedConfig = {
  embeddedSqlTags: ["sql"],
};

const customDialect = {
  ...postgresql,
  tokenizerOptions:{
    ...postgresql.tokenizerOptions,
    reservedDataTypes:[
      ...postgresql.tokenizerOptions.reservedDataTypes,
      "bytea"
    ]
  }
};

/** @type {import('prettier-plugin-sql').SqlBaseOptions} */
const prettierPluginSqlConfig = {
  formatter: "sql-formatter",
  // language: "postgresql",
  dialect: JSON.stringify(customDialect),
  keywordCase: "upper",
  functionCase: "lower",
  identifierCase: "lower",
  dataTypeCase: "upper",
  indentStyle: "standard",
  logicalOperatorNewline: "before",
};

const config = {
  ...prettierConfig,
  ...prettierPluginEmbedConfig,
  ...prettierPluginSqlConfig,
};

export default config;
