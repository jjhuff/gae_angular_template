var gulp = require('gulp');
var uglify = require('gulp-uglify');
var concat = require('gulp-concat');
var size = require('gulp-size');
//var clean = require('gulp-clean');
var rename = require('gulp-rename');
var minifyCSS = require('gulp-minify-css');
var minifyHTML = require('gulp-minify-html');
//var changed = require('gulp-changed');
var compass = require('gulp-compass');
var bowerFiles = require('main-bower-files');
var filter = require('gulp-filter');
var filelog = require('gulp-filelog');
var rev = require('gulp-rev');
var ngHtml2Js = require("gulp-ng-html2js");
var merge = require('merge-stream');
var lazypipe = require('lazypipe');
//var open = require('open');
//var runSequence = require('run-sequence');
//var connect = require('gulp-connect');
var shell = require('gulp-shell');

process.env.GOPATH = process.cwd();

var paths = {
    appjs: {
        src: './src/example.com/frontend/scripts/**/*.js',
        dest: './src/example.com/frontend/build/'
    },
    libjs: {
        dest: './src/example.com/frontend/build/'
    },
    views: {
        src: './src/example.com/frontend/views/*.html',
        base: './src/example.com/frontend/views/',
        dest: './src/example.com/frontend/build/'
    },
    styles: {
        src: './src/example.com/frontend/styles/*.scss',
        dest: './src/example.com/frontend/build/',
        sass: 'src/example.com/frontend/styles/',
        import_path: ['./src/example.com/frontend/vendor/bootstrap-sass-official/assets/stylesheets']
    },
    ae_extra: './src/example.com/frontend/',
    ae_yaml: {
        dispatch: ['./src/example.com/frontend/dispatch.yaml'],
        prod: ['./src/example.com/frontend/app-prod.yaml'],
        qa: ['./src/example.com/frontend/app-qa.yaml'],
        dev: ['./src/example.com/frontend/app-dev.yaml'],
    },
};

var options = {
  open: true,
  httpPort: 4400,
  devserver_port: 8080,
  admin_port: 8000
};

// Helper function to generate a pipe that does the common file revisions
function revFiles(name, dest) {
    return lazypipe()
        .pipe(size, {title: name})
        .pipe(rename, { suffix: '.min' })
        .pipe(rev)
        .pipe(gulp.dest, dest)
        .pipe(rev.manifest)
        .pipe(rename, name + "-manifest.json")
        .pipe(gulp.dest, dest);
}

// process the compass files
gulp.task('styles', function () {
    return gulp.src(paths.styles.src)
        .pipe(compass({
            css: paths.styles.dest,
            sass: paths.styles.sass,
            import_path: paths.styles.import_path
        }))
        .pipe(gulp.dest(paths.styles.dest))
        .pipe(minifyCSS())
        .pipe(size({title:"styles"}))
        .pipe(rename({ suffix: '.min' }))
        .pipe(gulp.dest(paths.styles.dest))
        .pipe(rev())
        .pipe(gulp.dest(paths.styles.dest))
        .pipe(rev.manifest())
        .pipe(rename("css-manifest.json"))
        .pipe(gulp.dest(paths.styles.dest));
});

gulp.task('minify-appjs', function () {
    var appjs = gulp.src(paths.appjs.src);
    var viewjs = gulp.src(paths.views.src)
        .pipe(minifyHTML({
            empty: true,
            quotes: true
        }))
        .pipe(ngHtml2Js({
            moduleName: "app",
            stripPrefix: paths.views.base,
            prefix: "/_/views/"
        }));

    return merge(appjs, viewjs)
        .pipe(concat("app.js"))
        .pipe(gulp.dest(paths.appjs.dest))
        .pipe(uglify())
        .pipe(revFiles("appjs", paths.appjs.dest)());
});

gulp.task('minify-libjs', function () {
    return gulp.src(bowerFiles()).pipe(filter([
            '**/*.js',
            '!bootstrap.js',
            '!angular.js',
            '!jquery.js',
        ]))
        //.pipe(filelog())
        .pipe(concat("lib.js"))
        .pipe(gulp.dest(paths.libjs.dest))
        .pipe(uglify())
        .pipe(revFiles("libjs", paths.libjs.dest)());
});

gulp.task('build', ['styles', 'minify-appjs', 'minify-libjs']);

gulp.task('watch', ['build'], function() {
    gulp.watch(paths.styles.src, ['styles']);
    gulp.watch(paths.views.src, ['minify-appjs']);
    gulp.watch(paths.appjs.src, ['minify-appjs']);
    gulp.watch('./bower.json', ['minify-libjs']);
});

gulp.task('server', ['watch'], function(){
    var cfg = [].concat(paths.ae_yaml.dev, paths.ae_yaml.dispatch);
    gulp.src('').pipe(shell([
        'goapp serve --port='+options.devserver_port+' --admin_port='+options.admin_port+' ' + cfg.join(' ')
    ]));
});

gulp.task('deploy:qa', ['build'], function(){
    var appcfg = 'appcfg.py --oauth2 -A example-qa ';
    gulp.src('').pipe(shell([
        appcfg + 'update_queues ' + paths.ae_extra,
        appcfg + 'update_cron ' + paths.ae_extra,
        appcfg + 'update ' + paths.ae_yaml.qa.join(' '),
        appcfg + 'update_dispatch ' + paths.ae_extra,
    ]));
});

gulp.task('deploy:prod', ['build'], function(){
    var appcfg = 'appcfg.py --oauth2 -A example-prod ';
    gulp.src('').pipe(shell([
        appcfg + 'update_queues ' + paths.ae_extra,
        appcfg + 'update_cron ' + paths.ae_extra,
        appcfg + 'update ' + paths.ae_yaml.prod.join(' '),
        appcfg + 'update_dispatch ' + paths.ae_extra,
    ]));
});

gulp.task('vet', shell.task([
  'go vet `go list example.com/... | grep -v third_party`',
],{ignoreErrors:true}));

gulp.task('goget', shell.task([
  'goapp get example.com/...',
], {ignoreErrors: true}));

gulp.task('test', shell.task([
  'goapp test -parallel 4  `go list example.com/... | grep -v third_party`',
],{ignoreErrors:true}));

gulp.task('install', ['goget'], function() {
  return bower.commands.install()
    .on('log', function(data) {
      gutil.log('bower', gutil.colors.cyan(data.id), data.message);
    });
});

// default gulp task
gulp.task('default', ['server'], function() {
});
