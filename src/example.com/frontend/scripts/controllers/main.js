'use strict';

angular.module('app')
  .controller('MainCtrl', ['$scope', '$modal', '$state', 'User', function ($scope, $modal, $state, User) {
        $scope.isLoggedIn = User.isLoggedIn;
  }]);
