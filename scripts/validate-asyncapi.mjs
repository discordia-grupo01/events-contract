import { Parser } from "@asyncapi/parser";
import fs from "node:fs";

const parser = new Parser();
const file = process.argv[2];
const { document, diagnostics } = await parser.parse(fs.readFileSync(file, "utf8"));
const errors = diagnostics.filter((d) => d.severity === 0);
for (const d of errors) console.error(d.code, d.message, (d.path || []).join("/"));
if (!document || errors.length > 0) process.exit(1);
console.log(`${file}: ${document.channels().all().length} canales, ${document.operations().all().length} operaciones`);
