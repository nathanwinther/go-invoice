var gulp = require('gulp');

gulp.task('sass', function() {
  var sass = require('gulp-sass');
  var css = require('gulp-clean-css');
  return gulp.src('./template/*.scss')
    .pipe(sass())
    .pipe(css())
    .pipe(gulp.dest('./template'));
});

gulp.task('watch', function() {
  gulp.watch('./template/*.scss', ['sass']);
});

gulp.task('default', ['sass']);

