var persona = angular.module("persona", []);

persona.factory('sessionService', ["$http", "$q", function($http, $q) {
    var _sessionKey = 'mtrack.session';
    var _cached;
    var _session = function() {
        var deferred = $q.defer();

        setTimeout(function() {
            if (typeof _cached !== 'undefined') {
                // TODO check expiration time.
                deferred.resolve(Object.create(_cached));
                return;
            }

            var cached = localStorage[_sessionKey];
            if (typeof cached !== 'undefined') { 
                _cached = JSON.parse(cached);
                deferred.resolve(Object.create(_cached));
                return;
            }

            deferred.reject({ status: 'unauthenticated' });
        }, 0);

        return deferred.promise;
    };
    var _store = function(session) {
        var deferred = $q.defer();

        setTimeout(function() {
            if (typeof session === 'object') {
                _cached = session;
                localStorage[_sessionKey] = JSON.stringify(session);
                deferred.resolve('success');
                console.log('local storage success');
            } else {
                console.log('local storage failure');
                deferred.reject('argument is not an object');
            }
        }, 0);

        return deferred.promise;
    };
    var _delete = function() {
        var deferred = $q.defer();

        setTimeout(function() {
            if (typeof _cached !== 'undefined') {
                _cached = undefined;
                delete localStorage[_sessionKey];
                deferred.resolve('logged out');
                return;
            }
            deferred.reject('already logged out');
        }, 0);

        return deferred.promise;
    };
    return {
        session: _session,
        store: _store,
        endSession: _delete
    };
}]);

persona.factory("personaService", ["$http", "$q", "sessionService", function($http, $q, sessionService) {
    return {
        verify: function () {
                    var deferred = $q.defer();
                    navigator.id.get(function(assertion) {
                        $http.post(config.persona.verifyUrl, { assertion: assertion }).
                            success(function(data) {
                                console.log('verify success:', data);
                                sessionService.store(data).then(
                                    function(result) {
                                        console.log('storage success:', result);
                                    },
                                    function(reason) {
                                        console.log('storage failure:', reason);
                                    });
                                deferred.resolve(data);
                            }).
                            error(function(data) {
                                console.log('verify failure:', data);
                                deferred.reject(data.reason);
                            });
                    });
                    return deferred.promise;
                },
        logout: function () {
                    if (config.persona.logoutUrl) {
                        var deferred = $q.defer();

                        $http.post(config.persona.logoutUrl).
                            success(function(data) {
                                sessionService.endSession().then(
                                    function(message) { deferred.resolve(message); },
                                    function(message) { deferred.reject(message); });
                            }).
                            error(function(data) { deferred.reject(data); });

                        return deferred.promise;
                    } else {
                        return sessionService.endSession();
                    }
                }
    };
}]);
