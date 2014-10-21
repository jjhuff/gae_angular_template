'use strict';

angular.module('app')
    .controller('HeaderCtrl', ['$scope', '$location', 'Config', 'User', function ($scope, $location, Config, User) {
        $scope.config = Config;
        $scope.isLoggedIn = User.isLoggedIn;
        $scope.user = User.get()
        $scope.signOut = function() {
            User.signOut().then(function() {
                $location.path('/');
            });
        } //signOut

        $scope.isActive = function (path) {
            return path === $location.path();
        };
        $scope.isActivePrefix = function (prefix) {
            return $location.path().indexOf(prefix) === 0;
        };

    }]);
