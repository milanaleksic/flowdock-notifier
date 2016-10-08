'use strict';

const exec = require('child_process').exec;

exports.handler = (event, context, callback) => {
    console.log("Event: " + event + ", stringified: " + JSON.stringify(event));
    const child = exec("./app '" + JSON.stringify(event) + "'", (error) => {
        callback(error, 'Process complete!');
    });

    child.stdout.on('data', console.log);
    child.stderr.on('data', console.error);
};
