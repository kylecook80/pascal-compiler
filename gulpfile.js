'use strict';

var gulp = require('gulp');
var gutil = require('gulp-util');
var exec = require('child_process').exec;

// var minimist = require('minimist');
// var options = minimist(process.argv);
// var environment = options.environment || 'development';

gulp.task('compile', function() {
    exec('sh ./final.sh');
});

gulp.task('watch', function() {
    gulp.watch('**/*.go', ['compile']);
});

gulp.task('default', ['compile', 'watch'])
