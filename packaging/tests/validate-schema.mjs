import fs from 'node:fs';
import Ajv2020 from 'ajv/dist/2020.js';
import addFormats from 'ajv-formats';

const [, , schemaPath, dataPath] = process.argv;
if (!schemaPath || !dataPath) {
  console.error('Usage: node validate-schema.mjs <schema> <data>');
  process.exit(2);
}

const ajv = new Ajv2020({ allErrors: true, strict: true });
addFormats(ajv);
const validate = ajv.compile(JSON.parse(fs.readFileSync(schemaPath, 'utf8')));
const valid = validate(JSON.parse(fs.readFileSync(dataPath, 'utf8')));
if (!valid) {
  console.error(JSON.stringify(validate.errors, null, 2));
  process.exit(1);
}
