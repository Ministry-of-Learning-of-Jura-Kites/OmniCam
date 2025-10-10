import { postgresql } from "sql-formatter";

/** @type {import('prettier').Config} */
const prettierConfig = {
  plugins: ["prettier-plugin-embed", "prettier-plugin-sql"],
};

/** @type {import('prettier-plugin-embed').PrettierPluginEmbedOptions} */
const prettierPluginEmbedConfig = {
  embeddedSqlTags: ["sql"],
};

/** @type {import('sql-formatter').DialectOptions} */
const customDialect = {
  ...postgresql,
  tokenizerOptions: {
    ...postgresql.tokenizerOptions,
    reservedDataTypes: [
      ...postgresql.tokenizerOptions.reservedDataTypes,
      "bytea",
    ],
    reservedFunctionNames: [
      ...postgresql.tokenizerOptions.reservedFunctionNames,
      "sqlc.embed",
      "sqlc.arg",
      "sqlc.narg",
    ],
  },
};

/** @type {import('prettier-plugin-sql').SqlBaseOptions} */
const prettierPluginSqlConfig = {
  formatter: "sql-formatter",
  // language: "postgresql",
  dialect: JSON.stringify(customDialect),
  keywordCase: "upper",
  functionCase: "upper",
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
