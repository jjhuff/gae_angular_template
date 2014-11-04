'use strict';

angular.module('app')
    .controller('SettingsCtrl', function ($scope, $location, User) {
        $scope.pass = {};
        $scope.current_email = User.get().email;
        $scope.promise = null;

        $scope.savePassword = function() {
            $scope.promise = User.changePassword($scope.pass.old, $scope.pass.new).then(function() {
                $scope.passwordForm.error = null;
                $scope.passwordForm.$setPristine();
                $scope.pass = {}
            }).catch(function(err) {
                $scope.passwordForm.error = err.status || "unknown";
            });
        }

        $scope.saveEmail = function() {
            User.get().email = $scope.email;
            $scope.promise = User.save().then(function() {
                $scope.emailForm.error = null;
                $scope.emailForm.$setPristine();
                $scope.email = '';
                $scope.email_conf = '';
                $scope.current_email = User.get().email;
            }).catch(function(err) {
                $scope.emailForm.error = err.status || "unknown";
            });
        }
    });
