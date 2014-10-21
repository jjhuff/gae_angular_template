'use strict';

angular.module('app')
    .controller('SignUpCtrl', ['$scope', '$location', 'User', 'Config', function ($scope, $location, User, Config) {
        $scope.promise = null;
        if (Config.showRegisterRandom) {
            $scope.randUser = function() {
                var rand = Math.round(Math.random()*10000).toString();
                $scope.email = "jjhuff+"+rand+"@mspin.net";
                $scope.email_conf = $scope.email;
                $scope.password = "1234567";
                $scope.password_conf = $scope.password;
            }
        }

        $scope.register = function() {
            $scope.promise = User.register({
                email: $scope.email,
                password: $scope.password
            }).then(function() {
                $location.path('/');
            }).catch(function(err) {
                $scope.error = err.status || 'unknown';
            });
        }; //register

    }]);
