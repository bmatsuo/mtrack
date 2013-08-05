var config = {
    persona: {
                 verifyUrl: '/api/persona/verify',
                 statusUrl: null,
                 logoutUrl: null
             }
};

var mtrack = angular.module('mtrack', ["persona"]);
