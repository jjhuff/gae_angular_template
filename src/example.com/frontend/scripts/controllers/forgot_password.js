'use strict';

angular.module('app')
    .controller('ResetPasswordCtrl', ['$scope', '$location', '$stateParams', 'User', function ($scope, $location, $stateParams, User) {
        $scope.promise = null;
        $scope.pass = {};

        $scope.savePassword = function() {
            $scope.promise = User.resetPassword($stateParams.token, $scope.pass.new).then(function() {
                $location.path('/');
            }).catch(function(err) {
                $scope.error = err.status;
            });
        }
    }])
    .controller('SendForgotPasswordCtrl', ['$scope', '$modalInstance', '$location', 'User', function($scope, $modalInstance, $location, User) {
        // Nest this down a level to avoid issues with nextes Angular scopes
        $scope.forgot = {email:$scope.$parent.email};
        $scope.ok = function () {
            User.sendForgotPassword($scope.forgot.email).then(function(){
                $modalInstance.dismiss('submit');
                $location.path('/');
            }).catch(function(err) {
                $scope.error = err.status;
            });
        };

        $scope.cancel = function () {
            $modalInstance.dismiss('cancel');
        };
    }]);
