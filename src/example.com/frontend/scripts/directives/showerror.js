'use strict';

var showErrorDirective = function () {
    return {
        restrict: 'E',
        require: '?ngModel',
        link: function (scope, elm, attr, ctrl) {
            if (!ctrl) {
                return;
            }

            scope.$watch(function() {
                return ctrl.hasVisited && !ctrl.hasFocus;
            }, function(v) {
                ctrl.showError = v;
            });
        }
    };
};

angular.module('app')
    .directive('input', showErrorDirective)
    .directive('select', showErrorDirective);
