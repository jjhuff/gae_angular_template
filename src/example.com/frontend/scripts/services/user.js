'use strict';

angular.module('app')
.factory('User', ['$http', function($http){

    var apiRoot = '/_/api/v1';

    var anonUser = { email: '' };
    var currentUser = _.clone(anonUser);

    function changeUser(user) {
        _.extend(currentUser, user);
    };

    // Set internal state based on the current user
    if (window.ct_user) {
        changeUser(ct_user);
        window.ct_user = undefined;
    }

    return {
        isLoggedIn: function() {
            return currentUser.email !== '';
        },
        register: function(user) {
            return $http.post(apiRoot+'/register', user).then(function(resp) {
                changeUser(resp.data);
            });
        },
        signIn: function(user) {
            return $http.post(apiRoot+'/login', user).then(function(resp) {
                changeUser(resp.data);
            });
        },
        signOut: function() {
            return $http.post(apiRoot+'/logout').then(function(){
                currentUser = _.clone(anonUser);
            });
        },
        changePassword: function(oldPassword, newPassword) {
            var d = {
                'old': oldPassword,
                'new': newPassword
            }
            return $http.post(apiRoot+'/change_password', d).then(function(resp){
                changeUser(resp.data);
            });
        },
        sendForgotPassword: function(email) {
            return $http.post(apiRoot+'/forgot_password', {'email':email})
        },
        resetPassword: function(token, newPassword) {
            var d = {
                'token': token,
                'new': newPassword
            }
            return $http.post(apiRoot+'/reset_password', d).then(function(resp){
                changeUser(resp.data);
            });
        },
        save: function() {
            return $http.put(apiRoot+'/user/'+currentUser.id, currentUser).then(function(resp){
                changeUser(resp.data);
            });
        },

        get: function() {
            return currentUser;
        }
    };
}]);

