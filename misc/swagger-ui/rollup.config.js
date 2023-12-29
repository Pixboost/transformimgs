import resolve from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import { terser } from 'rollup-plugin-terser';
import builtins from 'rollup-plugin-node-builtins'
import * as fs from "fs";
import css from "rollup-plugin-css-only";

// `npm run build` -> `production` is true
// `npm run dev` -> `production` is false
const production = !process.env.ROLLUP_WATCH;

export default {
  input: 'src/main.js',
  output: {
    file: 'public/bundle.js',
    format: 'iife', // immediately-invoked function expression — suitable for <script> tags
    sourcemap: true
  },
  plugins: [
    resolve(), // tells Rollup how to find date-fns in node_modules
    commonjs(), // converts date-fns to ES modules
    builtins(),
    production && terser(), // minify, but only in production
    css({output: 'bundle.css'})
  ]
};


fs.promises
  .copyFile(`${__dirname}/../../swagger.yaml`, `${__dirname}/public/swagger.yaml`)
  .then(() => {
    console.log('Successfully copied spec');
  })
  .catch(err => {
    console.err('Failed to copy swagger spec', err)
  });