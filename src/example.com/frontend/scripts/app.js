'use strict';

var app = angular.module('app', [
  'ngCookies',
  'ngResource',
  'ngSanitize',
  'ui.bootstrap',
  'ui.router',
  'angular-md5',
  'cgBusy'
]);

app.config(function ($stateProvider, $urlRouterProvider, $locationProvider) {
    $locationProvider.html5Mode(true).hashPrefix('!');
    $urlRouterProvider.otherwise("/");
    $stateProvider
    .state("main", {
        url: "/",
        templateUrl: "/_/views/main.html",
    })
    .state("signup", {
        url: "/signup",
        templateUrl: "/_/views/signup.html",
    })
    .state("login", {
        url: "/login",
        templateUrl: "/_/views/signin.html",
    })
    .state("reset_password", {
        url: "/reset_password/:token",
        templateUrl: "/_/views/reset_password.html",
    })
    .state("contact", {
        url: "/contact",
        templateUrl: "/_/views/contact.html",
    })
    .state("about", {
        url: "/about",
        templateUrl: "/_/views/about.html",
    })
    .state('settings', {
        url: "/settings",
        templateUrl: '/_/views/settings.html',
        loginRequired: true,
    })
});


// If a path is flagged as login required, require it
app.run(function ($rootScope, $state, User) {
    $rootScope.$on("$stateChangeStart", function (event, toState, toParams, fromState, fromParams) {
        if (toState.loginRequired && !User.isLoggedIn()) {
            event.preventDefault();
            $state.go("login");
        }
    });
});


// Intercept 401 errors and send to login page
app.config(function ($httpProvider) {
    $httpProvider.interceptors.push(function($q, $injector) {
        return {
            'response': function(response) {
                return response;
            },
            'responseError': function(response) {
                if (response.status === 401) {
                    $injector.get('$state').go('login');
                    // TODO: Do something to redirect back to $location.path
                }
                return $q.reject(response);
            }
        }
    });
});


app.filter('encodeUri', function ($window) {
    return $window.encodeURIComponent;
});

angular.module('app').value('cgBusyDefaults',{
    templateUrl: '/_/views/angular-busy.html',
    delay: 500,
    minDuration: 500,
});

// Process analytics pageviews
app.run(function ($rootScope, $location) {
    $rootScope.$on('$viewContentLoaded', function(){
        ga('send', 'pageview', $location.path());
    });
});
