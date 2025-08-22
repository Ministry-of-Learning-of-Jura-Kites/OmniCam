/** @type {import('prettier').Config} */
const prettierConfig = {
  plugins: ["prettier-plugin-embed", "prettier-plugin-sql"],
};

/** @type {import('prettier-plugin-embed').PrettierPluginEmbedOptions} */
const prettierPluginEmbedConfig = {
  embeddedSqlTags: ["sql"],
};

/** @type {import('prettier-plugin-sql').SqlBaseOptions} */
const prettierPluginSqlConfig = {
  formatter: "sql-formatter",
  language: "postgresql",
  keywordCase: "upper",
  functionCase: "lower",
  identifierCase: "lower",
  indentStyle: "standard",
  logicalOperatorNewline: "before",
};

const config = {
  ...prettierConfig,
  ...prettierPluginEmbedConfig,
  ...prettierPluginSqlConfig,
};

export default config;
