'use strict';

angular.module('app')
  .controller('MainCtrl', function ($scope, $modal, $state, User) {
        $scope.isLoggedIn = User.isLoggedIn;
  });
