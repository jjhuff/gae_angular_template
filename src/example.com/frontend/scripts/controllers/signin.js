'use strict';

angular.module('app')
    .controller('SignInCtrl', ['$scope', '$location', '$modal', 'User', function ($scope, $location, $modal, User) {

        $scope.loginPromise = null;

        $scope.signIn = function() {
            $scope.loginPromise = User.signIn({
                email: $scope.email,
                password: $scope.password
            }).then(function() {
                $location.path('/');
            }).catch(function(err) {
                $scope.error = err.status;
            });
        }; //signIn

        $scope.forgotPassword = function() {
            var modalInstance = $modal.open({
                templateUrl: '/_/views/forgot_password_req.html',
                controller: 'SendForgotPasswordCtrl',
                scope: $scope
            });
        };
    }]);
