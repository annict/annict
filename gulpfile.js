var autoprefixer = require("autoprefixer");
var gulp = require("gulp");
var postcss = require("gulp-postcss");
var gutil = require("gulp-util");
var watch = require('gulp-watch');
var precss = require("precss");
var webpack = require("webpack");

gulp.task("postcss", function() {
  var processors = [
    autoprefixer({browsers: ["last 1 version"]}),
    precss
  ];
  return gulp.src("./frontend/stylesheets/application.css")
    .pipe(postcss(processors))
    .pipe(gulp.dest("./public/assets/stylesheets"));
});

gulp.task("webpack", function(callback) {
  webpack({
    entry: "./frontend/javascripts/application.js",
    output: {
      path: __dirname,
      filename: "./public/assets/javascripts/application.js"
    },
    module: {
      loaders: [
        {
          test: /\.js$/,
          exclude: /node_modules/,
          loader: "babel",
          query: {
            presets: ["es2015"]
          }
        }
      ]
    }
  }, function(err, stats) {
    if (err) throw new gutil.PluginError("webpack", err);
    gutil.log("[webpack]", stats.toString());
    callback();
  });
});

gulp.task("watch", function() {
  gulp.watch("./frontend/stylesheets/**/*.css", ["postcss"]);
  gulp.watch("./frontend/javascripts/**/*.js", ["webpack"]);
});

gulp.task("default", ["watch"]);
